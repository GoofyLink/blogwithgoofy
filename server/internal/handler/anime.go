package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"blogwitgoofy/server/internal/model"
	"blogwitgoofy/server/pkg/response"
)

// ---------- 公开接口：动漫排行榜 ----------

func ListAnimes(c *gin.Context) {
	var animes []model.Anime
	if err := model.DB.Where("status = ?", model.StatusPublished).
		Order("rank ASC, id ASC").Find(&animes).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, animes)
}

// ---------- 管理接口 ----------

type animeForm struct {
	Title       string `json:"title" binding:"required,max=191"`
	Cover       string `json:"cover" binding:"max=500"`
	Category    string `json:"category" binding:"max=100"`
	Region      string `json:"region" binding:"max=100"`
	Episodes    string `json:"episodes" binding:"max=100"`
	PlayCount   int    `json:"playCount"`
	Description string `json:"description" binding:"max=1000"`
	Rank        int    `json:"rank"`
	Status      int    `json:"status"`
}

func AdminListAnimes(c *gin.Context) {
	var animes []model.Anime
	if err := model.DB.Order("rank ASC, id ASC").Find(&animes).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, animes)
}

func CreateAnime(c *gin.Context) {
	var form animeForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "动漫标题不能为空")
		return
	}
	anime := model.Anime{
		Title: form.Title, Cover: form.Cover, Category: form.Category, Region: form.Region,
		Episodes: form.Episodes, PlayCount: form.PlayCount, Description: form.Description,
		Rank: form.Rank, Status: form.Status,
	}
	if err := model.DB.Create(&anime).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, anime)
}

func UpdateAnime(c *gin.Context) {
	var anime model.Anime
	if err := model.DB.First(&anime, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "动漫不存在")
		return
	}
	var form animeForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "动漫标题不能为空")
		return
	}
	anime.Title = form.Title
	anime.Cover = form.Cover
	anime.Category = form.Category
	anime.Region = form.Region
	anime.Episodes = form.Episodes
	anime.PlayCount = form.PlayCount
	anime.Description = form.Description
	anime.Rank = form.Rank
	anime.Status = form.Status
	if err := model.DB.Save(&anime).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, anime)
}

func DeleteAnime(c *gin.Context) {
	var anime model.Anime
	if err := model.DB.First(&anime, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "动漫不存在")
		return
	}
	model.DB.Delete(&anime)
	response.OK(c, nil)
}

