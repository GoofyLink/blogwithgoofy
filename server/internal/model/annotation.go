package model

import "time"

// Annotation 学习文章的段落批注（学习笔记）
type Annotation struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	ArticleID      uint      `gorm:"index;not null" json:"articleId"`
	ParagraphIndex int       `gorm:"index" json:"paragraphIndex"` // 段落序号，从 0 开始
	Quote          string    `gorm:"size:1000" json:"quote"`     // 批注摘录的原文
	Note           string    `gorm:"type:text" json:"note"`      // 学习笔记内容
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
