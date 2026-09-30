package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"blogwitgoofy/server/internal/model"
	"blogwitgoofy/server/pkg/response"
)

// ---------- 分类 ----------

func ListCategories(c *gin.Context) {
	var cats []model.Category
	if err := model.DB.Model(&model.Category{}).
		Select("categories.id, categories.name, COUNT(articles.id) AS article_count").
		Joins("LEFT JOIN articles ON articles.category_id = categories.id AND articles.status = ?",
			model.StatusPublished).
		Group("categories.id").Order("categories.id").
		Scan(&cats).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, cats)
}

func CreateCategory(c *gin.Context) {
	var form struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "分类名不能为空")
		return
	}
	cat := model.Category{Name: form.Name}
	if err := model.DB.Create(&cat).Error; err != nil {
		response.BadRequest(c, "创建失败：分类名可能已存在")
		return
	}
	response.OK(c, cat)
}

func UpdateCategory(c *gin.Context) {
	var cat model.Category
	if err := model.DB.First(&cat, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "分类不存在")
		return
	}
	var form struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "分类名不能为空")
		return
	}
	cat.Name = form.Name
	if err := model.DB.Save(&cat).Error; err != nil {
		response.BadRequest(c, "保存失败：分类名可能已存在")
		return
	}
	response.OK(c, cat)
}

func DeleteCategory(c *gin.Context) {
	var cat model.Category
	if err := model.DB.First(&cat, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "分类不存在")
		return
	}
	var cnt int64
	model.DB.Model(&model.Article{}).Where("category_id = ?", cat.ID).Count(&cnt)
	if cnt > 0 {
		response.BadRequest(c, "该分类下还有文章，请先移动或删除这些文章")
		return
	}
	model.DB.Delete(&cat)
	response.OK(c, nil)
}

// ---------- 标签 ----------

func ListTags(c *gin.Context) {
	var tags []model.Tag
	if err := model.DB.Model(&model.Tag{}).
		Select("tags.id, tags.name, COUNT(article_tags.article_id) AS article_count").
		Joins("LEFT JOIN article_tags ON article_tags.tag_id = tags.id").
		Group("tags.id").Order("tags.id").
		Scan(&tags).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, tags)
}

func CreateTag(c *gin.Context) {
	var form struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "标签名不能为空")
		return
	}
	tag := model.Tag{Name: form.Name}
	if err := model.DB.Create(&tag).Error; err != nil {
		response.BadRequest(c, "创建失败：标签可能已存在")
		return
	}
	response.OK(c, tag)
}

func UpdateTag(c *gin.Context) {
	var tag model.Tag
	if err := model.DB.First(&tag, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "标签不存在")
		return
	}
	var form struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "标签名不能为空")
		return
	}
	tag.Name = form.Name
	if err := model.DB.Save(&tag).Error; err != nil {
		response.BadRequest(c, "保存失败：标签可能已存在")
		return
	}
	response.OK(c, tag)
}

func DeleteTag(c *gin.Context) {
	var tag model.Tag
	if err := model.DB.First(&tag, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "标签不存在")
		return
	}
	model.DB.Exec("DELETE FROM article_tags WHERE tag_id = ?", tag.ID)
	model.DB.Delete(&tag)
	response.OK(c, nil)
}
