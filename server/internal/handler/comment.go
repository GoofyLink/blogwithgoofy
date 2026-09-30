package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"blogwitgoofy/server/internal/model"
	"blogwitgoofy/server/pkg/response"
)

// ---------- 公开接口 ----------

func ListComments(c *gin.Context) {
	articleID := c.Param("id")
	var comments []model.Comment
	if err := model.DB.Where("article_id = ?", articleID).
		Order("id ASC").Find(&comments).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, comments)
}

func CreateComment(c *gin.Context) {
	articleID, _ := strconv.Atoi(c.Param("id"))
	var article model.Article
	if err := model.DB.First(&article, articleID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "文章不存在")
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
	comment := model.Comment{
		ArticleID: uint(articleID),
		Nickname:  form.Nickname,
		Email:     form.Email,
		Content:   form.Content,
	}
	if err := model.DB.Create(&comment).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, comment)
}

// ---------- 管理接口 ----------

func AdminListComments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	query := model.DB.Model(&model.Comment{}).
		Select("comments.*, articles.title AS article_title").
		Joins("LEFT JOIN articles ON articles.id = comments.article_id")
	if aid := c.Query("articleId"); aid != "" {
		query = query.Where("comments.article_id = ?", aid)
	}

	var total int64
	query.Count(&total)

	var comments []model.Comment
	if err := query.Order("comments.id DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&comments).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.Page(c, comments, total)
}

func DeleteComment(c *gin.Context) {
	var comment model.Comment
	if err := model.DB.First(&comment, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "评论不存在")
		return
	}
	model.DB.Delete(&comment)
	response.OK(c, nil)
}
