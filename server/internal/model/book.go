package model

import "time"

const (
	LangEN = "en"
	LangDE = "de"
)

// Book 学习模块的书（英/德语原版书）或小说专栏的书
type Book struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Title       string    `gorm:"size:191;not null;index" json:"title"` // 主书名
	Subtitle    string    `gorm:"size:191" json:"subtitle"`             // 副书名（原文书名/中文名）
	Author      string    `gorm:"size:191" json:"author"`
	Language    string    `gorm:"size:10;default:en;index" json:"language"` // en / de / zh
	Cover       string    `gorm:"size:500" json:"cover"`
	Description string    `gorm:"size:2000" json:"description"`
	Type        int       `gorm:"default:0;index" json:"type"` // 0 学习书房 / 1 小说专栏
	Status      int       `gorm:"default:1;index" json:"status"`
	Sort        int       `gorm:"default:0" json:"sort"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	ChapterCount int64 `gorm:"->" json:"chapterCount,omitempty"` // 列表联表统计
}

// Chapter 书的章节，Content 为纯文本（按空行分段）
type Chapter struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	BookID      uint      `gorm:"index;not null" json:"bookId"`
	Title       string    `gorm:"size:191;not null" json:"title"`
	Content     string    `gorm:"type:longtext" json:"content"`
	Translation string    `gorm:"type:longtext" json:"translation"` // 可选：整章中文对照
	Sort        int       `gorm:"default:0;index" json:"sort"`
	Status      int       `gorm:"default:1" json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
