package model

import "time"

// 数据表关系：一个 AnalyticsVisit（页面访问）会收到多个事件；
// AnalyticsReceipt 记住处理过的事件 ID；AnalyticsFact 保存可供仪表盘查询的统计量。
// GORM 根据结构体自动建表；primaryKey 是主键，index 是加快筛选的索引。
//
// AnalyticsVisit 保存一次页面访问，保留 90 天。ID 对应前端 viewId，
// Seconds 保存“已处理的最大累计停留”，用于对后续心跳计算增量。
type AnalyticsVisit struct {
	ID        string    `gorm:"size:36;primaryKey"` // 主键：访问表用 viewId，凭据表用事件 id，事实表用维度散列
	Visitor   string    `gorm:"size:64"`            // 访客标识；传输时是随机 UUID，落库时为 HMAC
	Session   string    `gorm:"size:64"`            // 会话标识；落库前与访客标识一起计算 HMAC
	Path      string    `gorm:"size:200"`           // 不带查询参数和锚点的前台路径
	Module    string    `gorm:"size:30"`            // 服务端确认的模块编码，例如 game
	Title     string    `gorm:"size:191"`           // 服务端查得的内容标题
	Source    string    `gorm:"size:255"`           // 入口来源；落库前提取域名或归类
	Device    string    `gorm:"size:20"`            // 根据 User-Agent 推测的电脑/手机/平板
	Admin     bool      // 是否为经认证的管理员测试访问
	Day       string    `gorm:"size:10;index"` // 访问首次接收时的北京时间日期
	Hour      int       // 访问首次接收时的北京时间小时，0～23
	Seconds   int       // 累计有效停留秒数；事实表中为归并后的总秒数
	CreatedAt time.Time `gorm:"index"` // 服务端创建时间，用于启用时间展示或过期清理
}

// AnalyticsFact 是按维度归并的事实记录，保留 365 天。
// 不是一人一行：同一访客换页面、小时、来源等会有多行，查询 UV 时必须 DISTINCT。
// ID 由维度散列生成；Day + Module 联合索引用于日期范围及模块筛选。
// Action 为空时 Count 表示 PV、Seconds 表示停留；非空时 Count 表示点击次数。
type AnalyticsFact struct {
	ID      string `gorm:"size:64;primaryKey"`                                           // 主键：访问表用 viewId，凭据表用事件 id，事实表用维度散列
	Day     string `gorm:"size:10;index:idx_analytics_day_module,priority:1" json:"day"` // 访问首次接收时的北京时间日期
	Hour    int    // 访问首次接收时的北京时间小时，0～23
	Module  string `gorm:"size:30;index:idx_analytics_day_module,priority:2"` // 服务端确认的模块编码，例如 game
	Path    string `gorm:"size:200"`                                          // 不带查询参数和锚点的前台路径
	Title   string `gorm:"size:191"`                                          // 服务端查得的内容标题
	Visitor string `gorm:"size:64"`                                           // 访客标识；传输时是随机 UUID，落库时为 HMAC
	Session string `gorm:"size:64"`                                           // 会话标识；落库前与访客标识一起计算 HMAC
	Source  string `gorm:"size:255"`                                          // 入口来源；落库前提取域名或归类
	Device  string `gorm:"size:20"`                                           // 根据 User-Agent 推测的电脑/手机/平板
	Admin   bool   // 是否为经认证的管理员测试访问
	Action  string `gorm:"size:30"`  // 点击操作类型，页面统计时为空
	Target  string `gorm:"size:255"` // 外链目标；落库只保留域名
	Count   int64  // 页面事实为 PV 次数，点击事实为点击次数
	Seconds int64  // 累计有效停留秒数；事实表中为归并后的总秒数
}

// AnalyticsReceipt 只保存事件 ID 和接收时间，保留 90 天。
// 同一事件重试时插入相同主键会被忽略，从而不会再次执行统计增量。
type AnalyticsReceipt struct {
	ID        string    `gorm:"size:36;primaryKey"` // 主键：访问表用 viewId，凭据表用事件 id，事实表用维度散列
	CreatedAt time.Time `gorm:"index"`              // 服务端创建时间，用于启用时间展示或过期清理
}

// AnalyticsState 只有 ID=1 的一条记录，由数据库初始化时 FirstOrCreate 创建。
// 再次启动不会重置 CreatedAt，因此它表示这套统计首次启用的时间。
type AnalyticsState struct {
	ID        uint      `gorm:"primaryKey"` // 主键：访问表用 viewId，凭据表用事件 id，事实表用维度散列
	CreatedAt time.Time // 服务端创建时间，用于启用时间展示或过期清理
}
