package handler

import (
	"blogwitgoofy/server/internal/model"
	"blogwitgoofy/server/pkg/response"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 本文件负责接收埋点，阅读时先看 CollectAnalytics（HTTP 入口），
// 再看 ingestAnalytics（事务内写入）和 addAnalyticsFact（汇总）。
// 仪表盘读取这些数据的逻辑在 analytics_dashboard.go。
var errAnalyticsInvalid = errors.New("invalid analytics event")

// 全站统计统一按北京时间划分日期，避免服务器部署在其他时区后口径变化。
var analyticsZone = time.FixedZone("Asia/Shanghai", 8*3600)
var uuidPattern = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
var detailPattern = regexp.MustCompile(`^/(article|category|tag|anime|art|game|game/post|learn/book|novel/book|learn/chapter|novel/chapter)/([1-9][0-9]*)$`)
var analyticsModules = map[string]string{"home": "首页", "essays": "随笔", "learn": "学习", "novel": "小说", "anime": "动漫", "game": "游戏", "art": "艺术", "ai": "AI", "archives": "归档", "links": "友链", "about": "关于", "page": "单页"}

// analyticsEvent 对应前端 JSON。json 标签指定传输字段名，不是数据库列定义。
// ID 是事件编号；ViewID 是页面访问编号：同一次访问会产生多条不同 ID 的事件。
type analyticsEvent struct {
	ID        string `json:"id"`        // 主键：访问表用 viewId，凭据表用事件 id，事实表用维度散列
	ViewID    string `json:"viewId"`    // 关联同一次页面访问的编号
	Visitor   string `json:"visitor"`   // 访客标识；传输时是随机 UUID，落库时为 HMAC
	Session   string `json:"session"`   // 会话标识；落库前与访客标识一起计算 HMAC
	Path      string `json:"path"`      // 不带查询参数和锚点的前台路径
	Source    string `json:"source"`    // 入口来源；落库前提取域名或归类
	Kind      string `json:"kind"`      // page_view、page_engagement 或 click
	Seconds   int    `json:"seconds"`   // 累计有效停留秒数；事实表中为归并后的总秒数
	Action    string `json:"action"`    // 点击操作类型，页面统计时为空
	Target    string `json:"target"`    // 外链目标；落库只保留域名
	StartedAt int64  `json:"startedAt"` // 客户端访问开始时间，单位为毫秒
}

// analyticsHash 用服务端密钥计算 HMAC：相同输入得到相同散列，仍可用于去重。
// 访客、会话落库前使用它转换；不保存浏览器发来的原始身份标识。
// 这不是可解密的加密；更换 JWT 密钥会改变散列结果，影响跨变更时段的 UV。
func analyticsHash(value string) string {
	key := "analytics"
	if Cfg != nil {
		key = Cfg.JWT.Secret
	}
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(value))
	return hex.EncodeToString(h.Sum(nil))
}

// analyticsHost 只接受 HTTP(S) 地址并提取域名，丢掉路径、端口、参数和片段。
// 例如 https://example.com/watch?id=1 最终只留下 example.com。
func analyticsHost(raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") {
		return ""
	}
	host := strings.ToLower(u.Hostname())
	if len(host) > 253 {
		return ""
	}
	return host
}

// analyticsSource 把入口归为直接访问、已知搜索引擎或外部站点。
// 用“相等或以 .域名 结尾”识别子域名，避免把 fakegoogle.com 当成 google.com。
func analyticsSource(raw, ownHost string) string {
	host := analyticsHost(raw)
	if host == "" || host == strings.Split(ownHost, ":")[0] {
		return "直接访问"
	}
	for _, engine := range []string{"google.com", "google.com.hk", "baidu.com", "bing.com", "sogou.com", "so.com", "duckduckgo.com"} {
		if host == engine || strings.HasSuffix(host, "."+engine) {
			return "搜索引擎：" + engine
		}
	}
	return host
}

// analyticsDevice 根据浏览器声明的 User-Agent 粗略判断设备，不是精确设备识别。
func analyticsDevice(ua string) string {
	ua = strings.ToLower(ua)
	if strings.Contains(ua, "ipad") || strings.Contains(ua, "tablet") || (strings.Contains(ua, "android") && !strings.Contains(ua, "mobile")) {
		return "平板"
	}
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "iphone") {
		return "手机"
	}
	return "电脑"
}

// resolveAnalyticsPage 是服务端页面白名单，返回可信的模块和标题。
// 不接受客户端自己提供的模块名；详情要查数据库确认存在且已发布。
// 游戏攻略和章节还要检查父级游戏/书籍是否发布；后台路径不会匹配。
func resolveAnalyticsPage(db *gorm.DB, path string) (string, string, error) {
	static := map[string]string{"/": "home", "/essays": "essays", "/learn": "learn", "/learn/words": "learn", "/learn/words/review": "learn", "/novel": "novel", "/anime": "anime", "/game": "game", "/art": "art", "/ai": "ai", "/archives": "archives", "/links": "links"}
	if m, ok := static[path]; ok {
		return m, analyticsModules[m], nil
	}
	if path == "/about" || strings.HasPrefix(path, "/p/") {
		slug := strings.TrimPrefix(path, "/p/")
		if path == "/about" {
			slug = "about"
		}
		if len(slug) > 100 || strings.ContainsAny(slug, "/?#") {
			return "", "", errAnalyticsInvalid
		}
		var row struct{ Title string }
		result := db.Table("pages").Select("title").Where("slug = ?", slug).Take(&row)
		m := "page"
		if path == "/about" {
			m = "about"
		}
		return m, row.Title, result.Error
	}
	match := detailPattern.FindStringSubmatch(path)
	if match == nil {
		return "", "", errAnalyticsInvalid
	}
	id, err := strconv.ParseUint(match[2], 10, 32)
	if err != nil {
		return "", "", errAnalyticsInvalid
	}
	kind := match[1]
	var row struct {
		Title string
		Type  int
	}
	var result *gorm.DB
	switch kind {
	case "article":
		// 同一个文章地址可能属于随笔或学习，依据数据库 type 分类。
		result = db.Table("articles").Select("title, type").Where("id = ? AND status = 1", id).Take(&row)
		kind = "essays"
		if row.Type == 1 {
			kind = "learn"
		}
	case "category", "tag":
		table := "categories"
		if kind == "tag" {
			table = "tags"
		}
		result = db.Table(table).Select("name AS title").Where("id = ?", id).Take(&row)
		kind = "essays"
	case "anime":
		result = db.Table("animes").Select("title").Where("id = ? AND status = 1", id).Take(&row)
	case "art":
		result = db.Table("artworks").Select("title").Where("id = ? AND status = 1", id).Take(&row)
	case "game":
		result = db.Table("games").Select("title").Where("id = ? AND status = 1", id).Take(&row)
	case "game/post":
		result = db.Table("game_posts p").Select("p.title").Joins("JOIN games g ON g.id=p.game_id AND g.status=1").Where("p.id = ? AND p.status=1", id).Take(&row)
		kind = "game"
	case "learn/book", "novel/book":
		// 书籍归属以数据库类型为准，不只看客户端传入的路由前缀。
		result = db.Table("books").Select("title,type").Where("id = ? AND status=1", id).Take(&row)
		kind = "learn"
		if row.Type == 1 {
			kind = "novel"
		}
	case "learn/chapter", "novel/chapter":
		result = db.Table("chapters c").Select("CONCAT(b.title, ' / ', c.title) AS title, b.type").Joins("JOIN books b ON b.id=c.book_id AND b.status=1").Where("c.id = ? AND c.status=1", id).Take(&row)
		kind = "learn"
		if row.Type == 1 {
			kind = "novel"
		}
	}
	if result == nil {
		return "", "", errAnalyticsInvalid
	}
	return kind, truncateAnalytics(row.Title, 191), result.Error
}

// 按 Unicode 字符数截取，避免按字节截断中文导致乱码。
func truncateAnalytics(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n])
	}
	return s
}

// 先校验 UUID、路径、事件类型、时间和点击分类，避免无效输入进入事务。
// 允许客户端时间最多快 1 分钟，访问开始时间最多早 24 小时；停留上限 4 小时。
// 这只是格式与合理性校验，不能证明客户端上报的行为一定真实发生。
func validAnalyticsEvent(e analyticsEvent, now time.Time) bool {
	age := now.UnixMilli() - e.StartedAt
	return uuidPattern.MatchString(e.ID) && uuidPattern.MatchString(e.ViewID) && uuidPattern.MatchString(e.Visitor) && uuidPattern.MatchString(e.Session) && len(e.Path) <= 200 && !strings.ContainsAny(e.Path, "?#") && age >= -60000 && age <= 86400000 && e.Seconds >= 0 && e.Seconds <= 14400 && (e.Kind == "page_view" || e.Kind == "page_engagement" || e.Kind == "click") && (e.Kind != "click" || e.Action == "ai_tool" || e.Action == "watch_link" || e.Action == "outbound")
}

// addAnalyticsFact 将一次增量写到统计事实表，而不是每次上报都插入一条明细。
// action 为空表示页面统计，此时 count 是 PV；非空表示点击，此时 count 是点击次数。
// 同日期/小时/页面/身份/来源/设备/管理员状态/操作/目标的记录合并为一行。
func addAnalyticsFact(tx *gorm.DB, v model.AnalyticsVisit, action, target string, count, seconds int64) error {
	f := model.AnalyticsFact{Day: v.Day, Hour: v.Hour, Module: v.Module, Path: v.Path, Title: v.Title, Visitor: v.Visitor, Session: v.Session, Source: v.Source, Device: v.Device, Admin: v.Admin, Action: action, Target: target, Count: count, Seconds: seconds}
	// 把上述维度拼成稳定主键；同一组维度会命中同一行。
	f.ID = analyticsHash(fmt.Sprintf("%s|%d|%s|%s|%s|%s|%s|%t|%s|%s", f.Day, f.Hour, f.Path, f.Visitor, f.Session, f.Source, f.Device, f.Admin, action, target))
	// 数据库原子执行“没有就新增、有则加 count/seconds”，避免先读再写丢失并发增量。
	return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoUpdates: clause.Assignments(map[string]interface{}{"count": gorm.Expr("count + ?", count), "seconds": gorm.Expr("seconds + ?", seconds), "title": v.Title})}).Create(&f).Error
}

// ingestAnalytics 必须在事务内调用；任何一步失败，去重凭据和统计增量一起回滚。
// 否则可能出现“事件 ID 已记住，但统计没有成功”，导致重试永远无法补回数据。
func ingestAnalytics(tx *gorm.DB, e analyticsEvent, admin bool, source, device string, now time.Time) error {
	// 第 1 层去重：事件 ID 是主键，重复插入忽略。已经处理过的事件直接返回。
	receipt := model.AnalyticsReceipt{ID: e.ID, CreatedAt: now}
	r := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&receipt)
	if r.Error != nil {
		return r.Error
	}
	if r.RowsAffected == 0 {
		return nil
	}
	// 会话散列同时包含访客标识，防止不同访客意外复用会话 ID 时被合并。
	visitor, session := analyticsHash(e.Visitor), analyticsHash(e.Visitor+":"+e.Session)
	// 第 2 层去重：按 viewId 查找访问。不同事件可以属于同一次访问，只增加一次 PV。
	var v model.AnalyticsVisit
	err := tx.First(&v, "id = ?", e.ViewID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		module, title, err := resolveAnalyticsPage(tx, e.Path)
		if err != nil {
			return err
		}
		// 用首次接收该访问的服务端时间决定日期/小时，后续停留和点击沿用这个归属。
		// 即使停留心跳先于 page_view 到达，也能创建访问，避免丢失 PV。
		v = model.AnalyticsVisit{ID: e.ViewID, Visitor: visitor, Session: session, Path: e.Path, Module: module, Title: title, Source: source, Device: device, Admin: admin, Day: now.In(analyticsZone).Format("2006-01-02"), Hour: now.In(analyticsZone).Hour(), CreatedAt: now}
		r = tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&v)
		if r.Error != nil {
			return r.Error
		}
		// 只有真正插入访问的请求才加 PV；并发请求命中同一个 viewId 时不能重复加。
		if r.RowsAffected > 0 {
			if err = addAnalyticsFact(tx, v, "", "", 1, 0); err != nil {
				return err
			}
		}
	} else if err != nil {
		return err
	}
	// 锁住这条访问直到事务结束，保证并发心跳按顺序读取和更新累计秒数。
	if err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&v, "id = ?", e.ViewID).Error; err != nil {
		return err
	}
	// 同一 viewId 的身份、路径和管理员状态必须一致，不能挪用其他访问的编号。
	if v.Visitor != visitor || v.Session != session || v.Path != e.Path || v.Admin != admin {
		return errAnalyticsInvalid
	}
	// 有效停留不应超过“开始到现在”的时间，额外留 5 秒容差。
	seconds := e.Seconds
	limit := int((now.UnixMilli()-e.StartedAt)/1000) + 5
	if limit < 0 {
		limit = 0
	}
	if seconds > limit {
		seconds = limit
	}
	// 累计值只增不减：已有 30 秒，收到 45 秒只增加 15；收到 20/30 秒不增加。
	// 因此即使心跳乱序到达，或者用不同事件 ID 重发，也不会重复累加时长。
	if seconds > v.Seconds {
		delta := seconds - v.Seconds
		if err = addAnalyticsFact(tx, v, "", "", 0, int64(delta)); err != nil {
			return err
		}
		if err = tx.Model(&v).Update("seconds", seconds).Error; err != nil {
			return err
		}
	}
	// 点击复用这次访问的身份和模块，但保存到 action 非空的独立事实行。
	// 同一个点击重试由前面的 receipt 去重；用户再次点击会产生新事件 ID。
	if e.Kind == "click" {
		target := analyticsHost(e.Target)
		if target == "" {
			return errAnalyticsInvalid
		}
		return addAnalyticsFact(tx, v, e.Action, target, 1, 0)
	}
	return nil
}

// 进程内限流表，键是 IP 的 HMAC，不持久化原始 IP。多实例之间不共享此限额。
// Gin 会并发处理请求，用 Mutex 保护 map，避免并发读写导致崩溃。
var analyticsRate = struct {
	sync.Mutex
	Items map[string]struct {
		At time.Time
		N  int
	}
}{Items: make(map[string]struct {
	At time.Time
	N  int
})}

// 每个来源在一分钟窗口内最多 120 次请求，并定期移除过期键以限制内存增长。
func allowAnalyticsRate(ip string, now time.Time) bool {
	analyticsRate.Lock()
	defer analyticsRate.Unlock()
	key := analyticsHash(ip)
	v := analyticsRate.Items[key]
	if now.Sub(v.At) > time.Minute {
		v.At = now
		v.N = 0
	}
	if v.N >= 120 {
		return false
	}
	v.N++
	if len(analyticsRate.Items) > 10000 {
		for k, item := range analyticsRate.Items {
			if now.Sub(item.At) > time.Minute {
				delete(analyticsRate.Items, k)
			}
		}
		if len(analyticsRate.Items) > 10000 {
			return false
		}
	}
	analyticsRate.Items[key] = v
	return true
}

// CollectAnalytics 是 POST /api/v1/analytics/events 的入口。
// 顺序：过滤常见机器人 → 限流 → 校验请求 → 验证管理员 → 事务写入 → 返回状态。
// 204 表示无需返回数据；被主动排除的流量也返回 204，让客户端停止重试。
func CollectAnalytics(c *gin.Context) {
	now := time.Now()
	ua := strings.ToLower(c.Request.UserAgent())
	if strings.Contains(ua, "bot") || strings.Contains(ua, "spider") || strings.Contains(ua, "crawler") {
		c.Status(204)
		return
	}
	if !allowAnalyticsRate(c.ClientIP(), now) {
		c.Status(http.StatusTooManyRequests)
		return
	}
	// 限制单次正文为 32 KB、事件为 1～10 条，避免超大请求占用资源。
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 32768)
	var body struct {
		Events       []analyticsEvent `json:"events"`
		Token        string           `json:"token"`
		IncludeAdmin bool             `json:"includeAdmin"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.Events) == 0 || len(body.Events) > 10 {
		response.BadRequest(c, "invalid analytics batch")
		return
	}
	// sendBeacon 不能自定义 Authorization 请求头，所以允许从正文取得令牌。
	// 仍然通过现有认证逻辑验证身份，令牌不写入统计表。
	if body.Token != "" && c.GetHeader("Authorization") == "" {
		c.Request.Header.Set("Authorization", "Bearer "+body.Token)
	}
	// includeAdmin 只是允许测试访问被记录，不是身份证明；管理员身份取决于有效令牌。
	admin := optionalUserID(c) > 0
	if admin && !body.IncludeAdmin {
		c.Status(204)
		return
	}
	for _, e := range body.Events {
		if !validAnalyticsEvent(e, now) {
			response.BadRequest(c, "invalid analytics event")
			return
		}
	}
	// 整批使用同一个事务：任何事件失败，整批回滚。重试不会留下半批重复数据。
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		for _, e := range body.Events {
			if err := ingestAnalytics(tx, e, admin, analyticsSource(e.Source, c.Request.Host), analyticsDevice(ua), now); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		// 无效事件返回 400，前端丢弃；数据库等临时故障返回 500，前端保留并重试。
		if errors.Is(err, errAnalyticsInvalid) || errors.Is(err, gorm.ErrRecordNotFound) {
			response.BadRequest(c, "analytics event not accepted")
		} else {
			response.Fail(c, http.StatusInternalServerError, "analytics temporarily unavailable")
		}
		return
	}
	c.Status(204)
}

// StartAnalyticsMaintenance 在服务启动时启动后台清理协程，不阻塞 HTTP 请求。
// 访问和事件凭据保留 90 天；统计事实保留含今天在内的 365 个北京时间日期。
// 事实表仍有散列访客/会话标识，以支持整个日期区间去重，不是只有每日总数。
func StartAnalyticsMaintenance() {
	go func() {
		for {
			now := time.Now()
			model.DB.Where("created_at < ?", now.AddDate(0, 0, -90)).Delete(&model.AnalyticsVisit{})
			model.DB.Where("created_at < ?", now.AddDate(0, 0, -90)).Delete(&model.AnalyticsReceipt{})
			model.DB.Where("day < ?", now.In(analyticsZone).AddDate(0, 0, -364).Format("2006-01-02")).Delete(&model.AnalyticsFact{})
			time.Sleep(6 * time.Hour)
		}
	}()
}
