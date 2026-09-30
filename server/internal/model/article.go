package model

import "time"

const (
	ArticleTypeBlog     = 0 // 普通博客
	ArticleTypeLearning = 1 // 英德语学习文章

	StatusDraft     = 0
	StatusPublished = 1
)

type Article struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Title       string    `gorm:"size:191;not null;index" json:"title"`
	Summary     string    `gorm:"size:1000" json:"summary"`
	Cover       string    `gorm:"size:500" json:"cover"`
	Content     string    `gorm:"type:longtext" json:"content"`
	Translation string    `gorm:"type:longtext" json:"translation"`
	CategoryID  uint      `gorm:"index" json:"categoryId"`
	Type        int       `gorm:"default:0;index" json:"type"`
	Language    string    `gorm:"size:10;default:zh;index" json:"language"` // zh / en / de
	Status      int       `gorm:"default:1;index" json:"status"`
	Views       int       `gorm:"default:0" json:"views"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Tags     []Tag     `gorm:"many2many:article_tags" json:"tags,omitempty"`
}
