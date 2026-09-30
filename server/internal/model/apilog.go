package model

import "time"

// ApiLog 接口调用日志
type ApiLog struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	Method     string    `gorm:"size:10;index" json:"method"`
	Path       string    `gorm:"size:500;index" json:"path"`
	IP         string    `gorm:"size:64;index" json:"ip"`
	Status     int       `json:"status"`
	DurationMs int       `json:"durationMs"` // 响应耗时（毫秒）
	UserAgent  string    `gorm:"size:500" json:"userAgent"`
	CreatedAt  time.Time `gorm:"index" json:"createdAt"`
}
