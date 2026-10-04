package model

import "time"

// ArtComment 艺术作品评论
type ArtComment struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	ArtID     uint      `gorm:"index;not null" json:"artId"`
	Nickname  string    `gorm:"size:100;not null" json:"nickname"`
	Email     string    `gorm:"size:191" json:"email"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"createdAt"`

	ArtTitle string `gorm:"->" json:"artTitle,omitempty"` // 后台联表填充
}
