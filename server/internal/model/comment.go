package model

import "time"

type Comment struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	ArticleID uint      `gorm:"index;not null" json:"articleId"`
	Nickname  string    `gorm:"size:100;not null" json:"nickname"`
	Email     string    `gorm:"size:191" json:"email"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"createdAt"`

	ArticleTitle string `gorm:"->" json:"articleTitle,omitempty"` // 后台列表联表填充
}
