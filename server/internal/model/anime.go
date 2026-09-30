package model

import "time"

// Anime 动漫排行榜条目
type Anime struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Title       string    `gorm:"size:191;not null" json:"title"`
	Cover       string    `gorm:"size:500" json:"cover"`         // 封面图片地址
	Category    string    `gorm:"size:100;index" json:"category"` // 分类：AI漫剧 / 漫画改 / 原创…（自由填写）
	Region      string    `gorm:"size:100" json:"region"`        // 地区：日本/中国/…
	Episodes    string    `gorm:"size:100" json:"episodes"`      // 更新至第X集 / 全12集
	PlayCount   int       `gorm:"default:0" json:"playCount"`    // 播放量（后台可改）
	Description string    `gorm:"size:1000" json:"description"`  // 一句话简介
	Rank        int       `gorm:"default:0;index" json:"rank"`   // 排名（小的在前）
	Status      int       `gorm:"default:1;index" json:"status"` // 1 上架 / 0 下架
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
