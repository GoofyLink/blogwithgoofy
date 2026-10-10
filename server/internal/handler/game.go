package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"blogwitgoofy/server/internal/model"
	"blogwitgoofy/server/pkg/response"
)

// ---------- 公开接口：游戏板块 ----------

func ListGames(c *gin.Context) {
	var games []model.Game
	q := model.DB.Where("status = ?", model.StatusPublished)
	if cat := c.Query("category"); cat != "" {
		q = q.Where("category = ?", cat)
	}
	if err := q.Select("games.*, (SELECT COUNT(*) FROM game_posts WHERE game_posts.game_id = games.id AND game_posts.status = 1) AS post_count").Order("sort ASC, id ASC").Find(&games).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, games)
}

// ---------- 管理接口 ----------

type gameForm struct {
	Content        string `json:"content" binding:"max=200000"`
	PlayStatus     string `json:"playStatus" binding:"omitempty,oneof=wishlist playing completed paused"`
	Review         string `json:"review" binding:"max=300"`
	Progress       string `json:"progress" binding:"max=200"`
	Recommended    bool   `json:"recommended"`
	RecommendOrder int    `json:"recommendOrder" binding:"min=0"`
	Title          string `json:"title" binding:"required,max=191"`
	Cover          string `json:"cover" binding:"max=500"`
	Category       string `json:"category" binding:"max=100"`
	Platform       string `json:"platform" binding:"max=100"`
	Description    string `json:"description" binding:"max=1000"`
	Tags           string `json:"tags" binding:"max=200"`
	Sort           int    `json:"sort"`
	Status         int    `json:"status" binding:"oneof=0 1"`
}

func AdminListGames(c *gin.Context) {
	var games []model.Game
	if err := model.DB.Order("sort ASC, id ASC").Find(&games).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, games)
}

func CreateGame(c *gin.Context) {
	var form gameForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "请检查游戏标题、字段长度和游玩状态")
		return
	}
	game := model.Game{
		Content: form.Content, PlayStatus: form.PlayStatus, Review: form.Review, Progress: form.Progress, Recommended: form.Recommended, RecommendOrder: form.RecommendOrder,
		Title: form.Title, Cover: form.Cover, Category: form.Category, Platform: form.Platform,
		Description: form.Description, Tags: form.Tags,
		Sort: form.Sort, Status: form.Status,
	}
	if err := model.DB.Create(&game).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, game)
}

func UpdateGame(c *gin.Context) {
	var game model.Game
	if err := model.DB.First(&game, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "游戏不存在")
		return
	}
	var form gameForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "请检查游戏标题、字段长度和游玩状态")
		return
	}
	game.Content = form.Content
	game.PlayStatus = form.PlayStatus
	game.Review = form.Review
	game.Progress = form.Progress
	game.Recommended = form.Recommended
	game.RecommendOrder = form.RecommendOrder
	game.Title = form.Title
	game.Cover = form.Cover
	game.Category = form.Category
	game.Platform = form.Platform
	game.Description = form.Description
	game.Tags = form.Tags
	game.Sort = form.Sort
	game.Status = form.Status
	if err := model.DB.Save(&game).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, game)
}

func DeleteGame(c *gin.Context) {
	var game model.Game
	if err := model.DB.First(&game, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "游戏不存在")
		return
	}
	model.DB.Delete(&game)
	response.OK(c, nil)
}
