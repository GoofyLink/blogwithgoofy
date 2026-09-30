package handler

import (
	"github.com/gin-gonic/gin"

	"blogwitgoofy/server/internal/model"
	"blogwitgoofy/server/pkg/response"
)

func Dashboard(c *gin.Context) {
	var articleCount, publishedCount, learningCount, commentCount,
		categoryCount, tagCount, viewTotal, bookCount, chapterCount int64

	model.DB.Model(&model.Article{}).Count(&articleCount)
	model.DB.Model(&model.Article{}).Where("status = ?", model.StatusPublished).Count(&publishedCount)
	model.DB.Model(&model.Article{}).Where("type = ?", model.ArticleTypeLearning).Count(&learningCount)
	model.DB.Model(&model.Comment{}).Count(&commentCount)
	model.DB.Model(&model.Category{}).Count(&categoryCount)
	model.DB.Model(&model.Tag{}).Count(&tagCount)
	model.DB.Model(&model.Article{}).Select("COALESCE(SUM(views), 0)").Scan(&viewTotal)
	model.DB.Model(&model.Book{}).Count(&bookCount)
	model.DB.Model(&model.Chapter{}).Count(&chapterCount)

	response.OK(c, gin.H{
		"articleCount":   articleCount,
		"publishedCount": publishedCount,
		"learningCount":  learningCount,
		"commentCount":   commentCount,
		"categoryCount":  categoryCount,
		"tagCount":       tagCount,
		"viewTotal":      viewTotal,
		"bookCount":      bookCount,
		"chapterCount":   chapterCount,
	})
}
