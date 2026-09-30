package model

import "time"

type Category struct {
	ID           uint   `gorm:"primarykey" json:"id"`
	Name         string `gorm:"size:100;uniqueIndex;not null" json:"name"`
	ArticleCount int64  `gorm:"->" json:"articleCount,omitempty"`
}

type Tag struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	Name         string    `gorm:"size:100;uniqueIndex;not null" json:"name"`
	ArticleCount int64     `gorm:"->" json:"articleCount,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}
