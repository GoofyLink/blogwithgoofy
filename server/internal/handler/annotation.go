package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"blogwitgoofy/server/internal/model"
	"blogwitgoofy/server/pkg/response"
)

// ---------- 公开接口：学习文章批注（读） ----------

func ListAnnotations(c *gin.Context) {
	var annotations []model.Annotation
	if err := model.DB.Where("article_id = ?", c.Param("id")).
		Order("paragraph_index ASC, id ASC").Find(&annotations).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, annotations)
}

// ---------- 管理接口：批注（增删改） ----------

type annotationForm struct {
	ArticleID      uint   `json:"articleId" binding:"required"`
	ParagraphIndex int    `json:"paragraphIndex"`
	Quote          string `json:"quote" binding:"max=1000"`
	Note           string `json:"note" binding:"required,max=10000"`
}

func CreateAnnotation(c *gin.Context) {
	var form annotationForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "批注内容不能为空")
		return
	}
	var article model.Article
	if err := model.DB.First(&article, form.ArticleID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "文章不存在")
		return
	}
	annotation := model.Annotation{
		ArticleID:      form.ArticleID,
		ParagraphIndex: form.ParagraphIndex,
		Quote:          form.Quote,
		Note:           form.Note,
	}
	if err := model.DB.Create(&annotation).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, annotation)
}

func UpdateAnnotation(c *gin.Context) {
	var annotation model.Annotation
	if err := model.DB.First(&annotation, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "批注不存在")
		return
	}
	var form struct {
		ParagraphIndex int    `json:"paragraphIndex"`
		Quote          string `json:"quote" binding:"max=1000"`
		Note           string `json:"note" binding:"required,max=10000"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "批注内容不能为空")
		return
	}
	annotation.ParagraphIndex = form.ParagraphIndex
	annotation.Quote = form.Quote
	annotation.Note = form.Note
	if err := model.DB.Save(&annotation).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, annotation)
}

func DeleteAnnotation(c *gin.Context) {
	var annotation model.Annotation
	if err := model.DB.First(&annotation, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "批注不存在")
		return
	}
	model.DB.Delete(&annotation)
	response.OK(c, nil)
}
