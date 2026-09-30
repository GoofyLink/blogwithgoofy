package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"blogwitgoofy/server/internal/model"
	"blogwitgoofy/server/pkg/response"
)

// ---------- 公开接口：AI 工具导航 ----------

func ListAiTools(c *gin.Context) {
	var tools []model.AiTool
	q := model.DB.Where("status = ?", model.StatusPublished)
	if s := c.Query("section"); s != "" {
		q = q.Where("section = ?", s)
	}
	if err := q.Order("sort ASC, id ASC").Find(&tools).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, tools)
}

func ListAiPrompts(c *gin.Context) {
	var prompts []model.AiPrompt
	q := model.DB.Where("status = ?", model.StatusPublished)
	if s := c.Query("section"); s != "" {
		q = q.Where("section = ?", s)
	}
	if err := q.Order("sort ASC, id ASC").Find(&prompts).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, prompts)
}

// ---------- 管理接口 ----------

type aiToolForm struct {
	Name        string `json:"name" binding:"required,max=100"`
	URL         string `json:"url" binding:"required,max=500"`
	Icon        string `json:"icon" binding:"max=500"`
	Section     string `json:"section" binding:"required,max=100"`
	Category    string `json:"category" binding:"max=100"`
	Description string `json:"description" binding:"max=1000"`
	Tags        string `json:"tags" binding:"max=200"`
	Sort        int    `json:"sort"`
	Status      int    `json:"status"`
}

func AdminListAiTools(c *gin.Context) {
	var tools []model.AiTool
	if err := model.DB.Order("section ASC, sort ASC, id ASC").Find(&tools).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, tools)
}

func CreateAiTool(c *gin.Context) {
	var form aiToolForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "名称和网址不能为空")
		return
	}
	tool := model.AiTool{
		Name: form.Name, URL: form.URL, Icon: form.Icon, Section: form.Section,
		Category: form.Category, Description: form.Description, Tags: form.Tags,
		Sort: form.Sort, Status: form.Status,
	}
	if err := model.DB.Create(&tool).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, tool)
}

func UpdateAiTool(c *gin.Context) {
	var tool model.AiTool
	if err := model.DB.First(&tool, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "工具不存在")
		return
	}
	var form aiToolForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "名称和网址不能为空")
		return
	}
	tool.Name = form.Name
	tool.URL = form.URL
	tool.Icon = form.Icon
	tool.Section = form.Section
	tool.Category = form.Category
	tool.Description = form.Description
	tool.Tags = form.Tags
	tool.Sort = form.Sort
	tool.Status = form.Status
	if err := model.DB.Save(&tool).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, tool)
}

func DeleteAiTool(c *gin.Context) {
	var tool model.AiTool
	if err := model.DB.First(&tool, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "工具不存在")
		return
	}
	model.DB.Delete(&tool)
	response.OK(c, nil)
}

// ---------- 提示词管理 ----------

type aiPromptForm struct {
	Title       string `json:"title" binding:"required,max=191"`
	Section     string `json:"section" binding:"required,max=100"`
	Category    string `json:"category" binding:"max=100"`
	Content     string `json:"content" binding:"required"`
	Description string `json:"description" binding:"max=500"`
	Sort        int    `json:"sort"`
	Status      int    `json:"status"`
}

func AdminListAiPrompts(c *gin.Context) {
	var prompts []model.AiPrompt
	if err := model.DB.Order("section ASC, sort ASC, id ASC").Find(&prompts).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, prompts)
}

func CreateAiPrompt(c *gin.Context) {
	var form aiPromptForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "标题和内容不能为空")
		return
	}
	p := model.AiPrompt{
		Title: form.Title, Section: form.Section, Category: form.Category,
		Content: form.Content, Description: form.Description,
		Sort: form.Sort, Status: form.Status,
	}
	if err := model.DB.Create(&p).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, p)
}

func UpdateAiPrompt(c *gin.Context) {
	var p model.AiPrompt
	if err := model.DB.First(&p, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "提示词不存在")
		return
	}
	var form aiPromptForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "标题和内容不能为空")
		return
	}
	p.Title = form.Title
	p.Section = form.Section
	p.Category = form.Category
	p.Content = form.Content
	p.Description = form.Description
	p.Sort = form.Sort
	p.Status = form.Status
	if err := model.DB.Save(&p).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, p)
}

func DeleteAiPrompt(c *gin.Context) {
	var p model.AiPrompt
	if err := model.DB.First(&p, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "提示词不存在")
		return
	}
	model.DB.Delete(&p)
	response.OK(c, nil)
}
