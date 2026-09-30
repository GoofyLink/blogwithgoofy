package model

import "time"

// Word 生词本：阅读时收藏的单词
type Word struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Word        string    `gorm:"size:191;uniqueIndex:idx_word_lang;not null" json:"word"`
	Language    string    `gorm:"size:10;uniqueIndex:idx_word_lang;not null" json:"language"` // en / de
	Translation string    `gorm:"size:500" json:"translation"`
	SourceText  string    `gorm:"size:500" json:"sourceText"` // 收藏时所在的句子（可选，便于回忆语境）
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
