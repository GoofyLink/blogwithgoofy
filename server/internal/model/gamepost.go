package model

import "time"

// GamePost 游戏攻略/资讯/评测帖子
type GamePost struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	GameID    uint      `gorm:"index;not null" json:"gameId"`
	Title     string    `gorm:"size:191;not null" json:"title"`
	Type      string    `gorm:"size:100;index" json:"type"` // 攻略 / 资讯 / 评测 / 心得
	Summary   string    `gorm:"size:1000" json:"summary"`
	Content   string    `gorm:"type:longtext" json:"content"`
	Views     int       `gorm:"default:0" json:"views"`
	Status    int       `gorm:"default:1;index" json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// GameComment 游戏圈评论（挂在帖子上）
type GameComment struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	PostID    uint      `gorm:"index;not null" json:"postId"`
	GameID    uint      `gorm:"index" json:"gameId"` // 冗余，后台列表展示用
	Nickname  string    `gorm:"size:100;not null" json:"nickname"`
	Email     string    `gorm:"size:191" json:"email"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"createdAt"`

	PostTitle string `gorm:"->" json:"postTitle,omitempty"` // 后台联表填充
	GameTitle string `gorm:"->" json:"gameTitle,omitempty"` // 后台联表填充
}
