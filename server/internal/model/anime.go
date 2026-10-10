package model

import "time"

// Anime 动漫排行榜条目
type Anime struct {
	WatchStatus       string     `gorm:"size:20;default:''" json:"watchStatus"`
	AirStatus         string     `gorm:"size:20;default:''" json:"airStatus"`
	Watched           int        `gorm:"default:0" json:"watched"`
	TotalEpisodes     *int       `json:"totalEpisodes"`
	Rating            *float64   `json:"rating"`
	Year              int        `gorm:"default:0" json:"year"`
	Source            string     `gorm:"size:100" json:"source"`
	Review            string     `gorm:"size:300" json:"review"`
	Content           string     `gorm:"type:longtext" json:"content"`
	Spoiler           bool       `gorm:"default:false" json:"spoiler"`
	Recommended       bool       `gorm:"default:false" json:"recommended"`
	RecommendOrder    int        `gorm:"default:0" json:"recommendOrder"`
	WatchURL          string     `gorm:"size:1000" json:"watchUrl"`
	WatchPlatform     string     `gorm:"size:100" json:"watchPlatform"`
	ProgressUpdatedAt *time.Time `json:"progressUpdatedAt"`
	ReviewUpdatedAt   *time.Time `json:"reviewUpdatedAt"`
	ID                uint       `gorm:"primarykey" json:"id"`
	Title             string     `gorm:"size:191;not null" json:"title"`
	Cover             string     `gorm:"size:500" json:"cover"`          // 封面图片地址
	Category          string     `gorm:"size:100;index" json:"category"` // 分类：AI漫剧 / 漫画改 / 原创…（自由填写）
	Region            string     `gorm:"size:100" json:"region"`         // 地区：日本/中国/…
	Episodes          string     `gorm:"size:100" json:"episodes"`       // 更新至第X集 / 全12集
	PlayCount         int        `gorm:"default:0" json:"playCount"`     // 播放量（后台可改）
	Description       string     `gorm:"size:1000" json:"description"`   // 一句话简介
	Rank              int        `gorm:"default:0;index" json:"rank"`    // 排名（小的在前）
	Status            int        `gorm:"index" json:"status"`            // 1 上架 / 0 下架
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}
