package model

import "time"

// Artwork 艺术鉴赏：展示的图片作品
type Artwork struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Title       string    `gorm:"size:191;not null" json:"title"`
	Image       string    `gorm:"size:500" json:"image"`        // 图片地址
	Description string    `gorm:"size:1000" json:"description"` // 简介
	Content     string    `gorm:"type:longtext" json:"content"` // 详情页详细介绍
	Author      string    `gorm:"size:191" json:"author"`       // 作者/来源（可选）
	Sort        int       `gorm:"default:0;index" json:"sort"`
	Status      int       `gorm:"default:1;index" json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
