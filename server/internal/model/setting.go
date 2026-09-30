package model

import "time"

// Setting 站点个性化配置（key-value），如头像、昵称、简介、联系方式
type Setting struct {
	Key       string    `gorm:"column:key;size:100;uniqueIndex;not null" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	UpdatedAt time.Time `json:"updatedAt"`
}
