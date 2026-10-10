package model

import "time"

// Game 游戏板块条目
type Game struct {
	PlayStatus     string    `gorm:"size:30;default:''" json:"playStatus"`
	Review         string    `gorm:"size:300" json:"review"`
	Progress       string    `gorm:"size:200" json:"progress"`
	Recommended    bool      `gorm:"default:false" json:"recommended"`
	RecommendOrder int       `gorm:"default:0" json:"recommendOrder"`
	PostCount      int64     `gorm:"->;-:migration" json:"postCount"`
	ID             uint      `gorm:"primarykey" json:"id"`
	Title          string    `gorm:"size:191;not null" json:"title"`
	Cover          string    `gorm:"size:500" json:"cover"`
	Category       string    `gorm:"size:100;index" json:"category"` // 电竞 / MOBA / FPS / 开放世界…
	Platform       string    `gorm:"size:100" json:"platform"`       // PC / 手机 / 主机 / 全平台
	Description    string    `gorm:"size:1000" json:"description"`
	Content        string    `gorm:"type:longtext" json:"content"` // 详情页详细介绍
	Tags           string    `gorm:"size:200" json:"tags"`         // 逗号分隔：免费,联机
	Sort           int       `gorm:"default:0;index" json:"sort"`
	Status         int       `gorm:"index" json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
