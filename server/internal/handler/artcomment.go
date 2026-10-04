package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"blogwitgoofy/server/internal/model"
	"blogwitgoofy/server/pkg/response"
)

// ---------- 公开接口 ----------

// GetArt 作品详情（浏览量暂不计，艺术以欣赏为主）
func GetArt(c *gin.Context) {
	var art model.Artwork
	if err := model.DB.Where("status = ?", model.StatusPublished).
		First(&art, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "作品不存在")
		return
	}
	response.OK(c, art)
}

func ListArtComments(c *gin.Context) {
	var comments []model.ArtComment
	if err := model.DB.Where("art_id = ?", c.Param("id")).
		Order("id ASC").Find(&comments).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, comments)
}

func CreateArtComment(c *gin.Context) {
	artID, _ := strconv.Atoi(c.Param("id"))
	var art model.Artwork
	if err := model.DB.First(&art, artID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "作品不存在")
		return
	}
	var form struct {
		Nickname string `json:"nickname" binding:"required,max=100"`
		Email    string `json:"email" binding:"omitempty,email,max=191"`
		Content  string `json:"content" binding:"required,max=2000"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "请填写昵称和评论内容（邮箱格式需正确）")
		return
	}
	comment := model.ArtComment{
		ArtID:    uint(artID),
		Nickname: strings.TrimSpace(form.Nickname),
		Email:    form.Email,
		Content:  form.Content,
	}
	if err := model.DB.Create(&comment).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, comment)
}

// ---------- 管理接口：评论 ----------

func AdminListArtComments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	query := model.DB.Model(&model.ArtComment{}).
		Select("art_comments.*, artworks.title AS art_title").
		Joins("LEFT JOIN artworks ON artworks.id = art_comments.art_id")

	var total int64
	query.Count(&total)
	var comments []model.ArtComment
	if err := query.Order("art_comments.id DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&comments).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.Page(c, comments, total)
}

func DeleteArtComment(c *gin.Context) {
	var comment model.ArtComment
	if err := model.DB.First(&comment, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "评论不存在")
		return
	}
	model.DB.Delete(&comment)
	response.OK(c, nil)
}
