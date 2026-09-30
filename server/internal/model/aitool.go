package model

import "time"

// AiTool AI 工具导航条目
type AiTool struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	URL         string    `gorm:"size:500;not null" json:"url"`
	Icon        string    `gorm:"size:500" json:"icon"`        // 图标/logo 地址（可空显示首字母）
	Section     string    `gorm:"size:100;index" json:"section"` // 一级大类：AI编程 / AI绘画 / AI写作…
	Category    string    `gorm:"size:100;index" json:"category"` // 二级分类：代码补全 / AI IDE / Chatbots…
	Description string    `gorm:"size:1000" json:"description"`
	Tags        string    `gorm:"size:200" json:"tags"` // 逗号分隔：免费,开源
	Sort        int       `gorm:"default:0;index" json:"sort"`
	Status      int       `gorm:"default:1;index" json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// AiPrompt AI 提示词收藏
type AiPrompt struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Title       string    `gorm:"size:191;not null" json:"title"`
	Section     string    `gorm:"size:100;index" json:"section"` // 同 AiTool.Section
	Category    string    `gorm:"size:100" json:"category"`
	Content     string    `gorm:"type:text" json:"content"`
	Description string    `gorm:"size:500" json:"description"`
	Sort        int       `gorm:"default:0;index" json:"sort"`
	Status      int       `gorm:"default:1;index" json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
