package model

import "time"

// ChapterNote 阅读时挂在段落上的个人笔记
type ChapterNote struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	ChapterID      uint      `gorm:"index;not null" json:"chapterId"`
	ParagraphIndex int       `gorm:"index" json:"paragraphIndex"` // 段落序号，从 0 开始
	Quote          string    `gorm:"size:1000" json:"quote"`     // 摘录原文（可选）
	Note           string    `gorm:"type:text" json:"note"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
