package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"blogwitgoofy/server/internal/model"
	"blogwitgoofy/server/pkg/response"
)

// ---------- 公开接口：艺术鉴赏 ----------

func ListArts(c *gin.Context) {
	var arts []model.Artwork
	if err := model.DB.Where("status = ?", model.StatusPublished).
		Order("sort ASC, id DESC").Find(&arts).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, arts)
}

// ---------- 管理接口 ----------

type artForm struct {
	Title       string `json:"title" binding:"required,max=191"`
	Image       string `json:"image" binding:"max=500"`
	Description string `json:"description" binding:"max=1000"`
	Content     string `json:"content"`
	Author      string `json:"author" binding:"max=191"`
	Sort        int    `json:"sort"`
	Status      int    `json:"status"`
}

func AdminListArts(c *gin.Context) {
	var arts []model.Artwork
	if err := model.DB.Order("sort ASC, id DESC").Find(&arts).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, arts)
}

func CreateArt(c *gin.Context) {
	var form artForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "作品标题不能为空")
		return
	}
	art := model.Artwork{
		Title: form.Title, Image: form.Image, Description: form.Description,
		Content: form.Content, Author: form.Author, Sort: form.Sort, Status: form.Status,
	}
	if err := model.DB.Create(&art).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, art)
}

func UpdateArt(c *gin.Context) {
	var art model.Artwork
	if err := model.DB.First(&art, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "作品不存在")
		return
	}
	var form artForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "作品标题不能为空")
		return
	}
	art.Title = form.Title
	art.Image = form.Image
	art.Description = form.Description
	art.Content = form.Content
	art.Author = form.Author
	art.Sort = form.Sort
	art.Status = form.Status
	if err := model.DB.Save(&art).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, art)
}

func DeleteArt(c *gin.Context) {
	var art model.Artwork
	if err := model.DB.First(&art, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "作品不存在")
		return
	}
	model.DB.Where("art_id = ?", art.ID).Delete(&model.ArtComment{})
	model.DB.Delete(&art)
	response.OK(c, nil)
}
