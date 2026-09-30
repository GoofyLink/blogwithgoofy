package model

import "time"

// Link 友情链接
type Link struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	Name        string `gorm:"size:100;not null" json:"name"`
	URL         string `gorm:"size:500;not null" json:"url"`
	Logo        string `gorm:"size:500" json:"logo"`
	Description string `gorm:"size:500" json:"description"`
	Sort        int    `gorm:"default:0" json:"sort"`
}

// Page 自定义单页（slug=about 即关于页）
type Page struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Slug      string    `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	Title     string    `gorm:"size:191" json:"title"`
	Content   string    `gorm:"type:longtext" json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
