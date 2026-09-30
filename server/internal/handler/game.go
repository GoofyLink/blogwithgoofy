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
	if err := q.Order("sort ASC, id ASC").Find(&games).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, games)
}

// ---------- 管理接口 ----------

type gameForm struct {
	Title       string `json:"title" binding:"required,max=191"`
	Cover       string `json:"cover" binding:"max=500"`
	Category    string `json:"category" binding:"max=100"`
	Platform    string `json:"platform" binding:"max=100"`
	Description string `json:"description" binding:"max=1000"`
	Tags        string `json:"tags" binding:"max=200"`
	Hot         int    `json:"hot"`
	Sort        int    `json:"sort"`
	Status      int    `json:"status"`
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
		response.BadRequest(c, "游戏标题不能为空")
		return
	}
	game := model.Game{
		Title: form.Title, Cover: form.Cover, Category: form.Category, Platform: form.Platform,
		Description: form.Description, Tags: form.Tags, Hot: form.Hot,
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
		response.BadRequest(c, "游戏标题不能为空")
		return
	}
	game.Title = form.Title
	game.Cover = form.Cover
	game.Category = form.Category
	game.Platform = form.Platform
	game.Description = form.Description
	game.Tags = form.Tags
	game.Hot = form.Hot
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
