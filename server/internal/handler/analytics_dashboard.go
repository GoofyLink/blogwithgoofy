package handler

import (
	"blogwitgoofy/server/internal/model"
	"blogwitgoofy/server/pkg/response"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

// 仪表盘顶部的整体指标。Seconds 是有效停留总秒数，AverageSeconds = Seconds / PV。
type analyticsMetrics struct {
	PV             int64   `json:"pv"`
	UV             int64   `json:"uv"`
	Sessions       int64   `json:"sessions"`
	Seconds        int64   `json:"seconds"`
	AverageSeconds float64 `json:"averageSeconds"`
}

// 通用排行行：Key 可以是模块、路径、来源或设备；Share 是 PV 占比（0～100）。
type analyticsGroup struct {
	Key            string  `json:"key"`
	Title          string  `json:"title"`
	PV             int64   `json:"pv"`
	UV             int64   `json:"uv"`
	Seconds        int64   `json:"seconds"`
	AverageSeconds float64 `json:"averageSeconds"`
	Share          float64 `json:"share"`
}

// 一个趋势点：单日查询时 Time 为小时，多日查询时为日期。
type analyticsTrend struct {
	Time string `json:"time"`
	PV   int64  `json:"pv"`
	UV   int64  `json:"uv"`
}

// SUM(count) 得到 PV，COUNT(DISTINCT visitor) 在整个查询范围去重得到 UV。
// 不能把每天 UV 相加：同一访客连续来 3 天，3 天的整体 UV 仍然只算 1。
// COALESCE 把空表聚合产生的 NULL 转成 0，方便前端展示。
const metricSelect = "COALESCE(SUM(count),0) AS pv, COUNT(DISTINCT visitor) AS uv, COUNT(DISTINCT session) AS sessions, COALESCE(SUM(seconds),0) AS seconds"

// 日期起止都包含当天，默认近 7 天。只允许最近 365 天内的范围，与数据保留期一致。
func analyticsDates(start, end string, now time.Time) (time.Time, time.Time, error) {
	today := now.In(analyticsZone)
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, analyticsZone)
	if start == "" {
		start = today.AddDate(0, 0, -6).Format("2006-01-02")
	}
	if end == "" {
		end = today.Format("2006-01-02")
	}
	a, e1 := time.ParseInLocation("2006-01-02", start, analyticsZone)
	b, e2 := time.ParseInLocation("2006-01-02", end, analyticsZone)
	if e1 != nil || e2 != nil || a.After(b) || b.After(today) || a.Before(today.AddDate(0, 0, -364)) {
		return a, b, fmt.Errorf("请选择最近365天内的有效日期范围")
	}
	return a, b, nil
}

// AnalyticsDashboard 对应受管理员认证保护的 GET /admin/analytics。
// 从统计事实表读取，依次返回整体指标、分组排行、补零趋势、点击排行和元信息。
func AnalyticsDashboard(c *gin.Context) {
	start, end, err := analyticsDates(c.Query("start"), c.Query("end"), time.Now())
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	module := c.Query("module")
	if module != "" && analyticsModules[module] == "" {
		response.BadRequest(c, "invalid module")
		return
	}
	// 每次调用都创建独立查询，共享相同的日期/模块/管理员筛选口径。
	// 不直接复用已追加 Group/Where 的查询对象，避免后一个查询带上前一个的条件。
	base := func() *gorm.DB {
		q := model.DB.Model(&model.AnalyticsFact{}).Where("day >= ? AND day <= ?", start.Format("2006-01-02"), end.Format("2006-01-02"))
		if c.Query("includeAdmin") != "true" {
			q = q.Where("admin = ?", false)
		}
		if module != "" {
			q = q.Where("module = ?", module)
		}
		return q
	}
	// action 为空才是页面记录，点击次数不能混进 PV 或页面停留指标。
	var metrics analyticsMetrics
	if err = base().Where("action = ''").Select(metricSelect).Scan(&metrics).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	if metrics.PV > 0 {
		metrics.AverageSeconds = float64(metrics.Seconds) / float64(metrics.PV)
	}
	// column 只来自下方写死的 module/path/source/device，不拼接用户输入作为列名。
	// 按各自分组去重访客；同一个访客可出现在多个模块，模块 UV 不能直接相加。
	// MAX(title) 选组内一个标题满足 SQL 聚合要求，并不代表最近一次标题。
	group := func(column string) ([]analyticsGroup, error) {
		rows := []analyticsGroup{}
		err := base().Where("action = ''").Select(column + " AS `key`, MAX(title) AS title, SUM(count) AS pv, COUNT(DISTINCT visitor) AS uv, SUM(seconds) AS seconds").Group(column).Order("uv DESC, pv DESC").Limit(100).Scan(&rows).Error
		for i := range rows {
			if rows[i].PV > 0 {
				rows[i].AverageSeconds = float64(rows[i].Seconds) / float64(rows[i].PV)
			}
			if metrics.PV > 0 {
				// 分母是当前筛选范围的总 PV；钻取单个模块后占比也跟着该范围变化。
				rows[i].Share = float64(rows[i].PV) / float64(metrics.PV) * 100
			}
		}
		return rows, err
	}
	modules, err := group("module")
	if err != nil {
		response.ServerError(c, err)
		return
	}
	for i := range modules {
		modules[i].Title = analyticsModules[modules[i].Key]
	}
	pages, err := group("path")
	if err != nil {
		response.ServerError(c, err)
		return
	}
	sources, err := group("source")
	if err != nil {
		response.ServerError(c, err)
		return
	}
	devices, err := group("device")
	if err != nil {
		response.ServerError(c, err)
		return
	}
	// 只选一天时按 00～23 小时聚合；多天则按日期聚合，避免图表过密。
	trend := []analyticsTrend{}
	hourly := start.Equal(end)
	key := "day"
	if hourly {
		key = "LPAD(hour,2,'0')"
	}
	if err = base().Where("action = ''").Select(key + " AS time, SUM(count) AS pv, COUNT(DISTINCT visitor) AS uv").Group(key).Order("time ASC").Scan(&trend).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	// 数据库只返回有访问的时段。先建索引，再遍历全部时段补 0，图表才不会缺点。
	indexed := map[string]analyticsTrend{}
	for _, r := range trend {
		indexed[r.Time] = r
	}
	trend = []analyticsTrend{}
	if hourly {
		for i := 0; i < 24; i++ {
			s := fmt.Sprintf("%02d", i)
			r := indexed[s]
			r.Time = s + ":00"
			trend = append(trend, r)
		}
	} else {
		for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
			s := d.Format("2006-01-02")
			r := indexed[s]
			r.Time = s
			trend = append(trend, r)
		}
	}
	// 点击独立聚合：按“操作类型 + 目标域名”统计次数及去重访客，最多返回 30 项。
	actions := []struct {
		Action string `json:"action"`
		Target string `json:"target"`
		Count  int64  `json:"count"`
		UV     int64  `json:"uv"`
	}{}
	if err = base().Where("action <> ''").Select("action, target, SUM(count) AS count, COUNT(DISTINCT visitor) AS uv").Group("action, target").Order("count DESC").Limit(30).Scan(&actions).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	// 单例记录用于展示统计启用时间，提醒用户旧浏览量没有回填。
	var state model.AnalyticsState
	model.DB.First(&state, 1)
	response.OK(c, gin.H{"metrics": metrics, "trend": trend, "modules": modules, "pages": pages, "sources": sources, "devices": devices, "actions": actions, "startedAt": state.CreatedAt, "timezone": "Asia/Shanghai", "retentionDays": 365})
}
