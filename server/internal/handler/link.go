package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"blogwitgoofy/server/internal/model"
	"blogwitgoofy/server/pkg/response"
)

// ---------- 友链 ----------

func ListLinks(c *gin.Context) {
	var links []model.Link
	if err := model.DB.Order("sort ASC, id ASC").Find(&links).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, links)
}

type linkForm struct {
	Name        string `json:"name" binding:"required,max=100"`
	URL         string `json:"url" binding:"required,max=500"`
	Logo        string `json:"logo"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
}

func CreateLink(c *gin.Context) {
	var form linkForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "友链名称和地址不能为空")
		return
	}
	link := model.Link{Name: form.Name, URL: form.URL, Logo: form.Logo,
		Description: form.Description, Sort: form.Sort}
	if err := model.DB.Create(&link).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, link)
}

func UpdateLink(c *gin.Context) {
	var link model.Link
	if err := model.DB.First(&link, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "友链不存在")
		return
	}
	var form linkForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "友链名称和地址不能为空")
		return
	}
	link.Name = form.Name
	link.URL = form.URL
	link.Logo = form.Logo
	link.Description = form.Description
	link.Sort = form.Sort
	if err := model.DB.Save(&link).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, link)
}

func DeleteLink(c *gin.Context) {
	var link model.Link
	if err := model.DB.First(&link, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "友链不存在")
		return
	}
	model.DB.Delete(&link)
	response.OK(c, nil)
}

// ---------- 单页 ----------

func GetPage(c *gin.Context) {
	var page model.Page
	if err := model.DB.Where("slug = ?", c.Param("slug")).First(&page).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "页面不存在")
		return
	}
	response.OK(c, page)
}

func AdminListPages(c *gin.Context) {
	var pages []model.Page
	if err := model.DB.Order("id ASC").Find(&pages).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, pages)
}

type pageForm struct {
	Slug    string `json:"slug" binding:"required,max=100"`
	Title   string `json:"title" binding:"required,max=191"`
	Content string `json:"content"`
}

func CreatePage(c *gin.Context) {
	var form pageForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "页面 slug 和标题不能为空")
		return
	}
	page := model.Page{Slug: form.Slug, Title: form.Title, Content: form.Content}
	if err := model.DB.Create(&page).Error; err != nil {
		response.BadRequest(c, "创建失败：slug 可能已被使用")
		return
	}
	response.OK(c, page)
}

func UpdatePage(c *gin.Context) {
	var page model.Page
	if err := model.DB.First(&page, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "页面不存在")
		return
	}
	var form pageForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "页面 slug 和标题不能为空")
		return
	}
	page.Slug = form.Slug
	page.Title = form.Title
	page.Content = form.Content
	if err := model.DB.Save(&page).Error; err != nil {
		response.BadRequest(c, "保存失败：slug 可能已被使用")
		return
	}
	response.OK(c, page)
}

func DeletePage(c *gin.Context) {
	var page model.Page
	if err := model.DB.First(&page, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "页面不存在")
		return
	}
	model.DB.Delete(&page)
	response.OK(c, nil)
}
