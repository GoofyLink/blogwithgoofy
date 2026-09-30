package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"blogwitgoofy/server/internal/model"
	"blogwitgoofy/server/pkg/response"
)

// ---------- 公开接口 ----------

func ListArticles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 50 {
		size = 10
	}

	query := model.DB.Model(&model.Article{}).
		Preload("Category").Preload("Tags").
		Where("status = ?", model.StatusPublished)

	if t := c.Query("type"); t != "" {
		query = query.Where("type = ?", t)
	}
	if lang := c.Query("language"); lang != "" {
		query = query.Where("language = ?", lang)
	}
	if cid := c.Query("categoryId"); cid != "" {
		query = query.Where("category_id = ?", cid)
	}
	if tid := c.Query("tagId"); tid != "" {
		query = query.Where("id IN (SELECT article_id FROM article_tags WHERE tag_id = ?)", tid)
	}
	if kw := strings.TrimSpace(c.Query("keyword")); kw != "" {
		like := "%" + kw + "%"
		query = query.Where("title LIKE ? OR summary LIKE ? OR content LIKE ?", like, like, like)
	}

	var total int64
	query.Count(&total)

	var articles []model.Article
	// 列表不取正文/翻译，减小响应体积
	err := query.Select("id, title, summary, cover, category_id, type, language, status, views, created_at, updated_at").
		Order("created_at DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&articles).Error
	if err != nil {
		response.ServerError(c, err)
		return
	}
	response.Page(c, articles, total)
}

func GetArticle(c *gin.Context) {
	id := c.Param("id")
	var article model.Article
	if err := model.DB.Preload("Category").Preload("Tags").First(&article, id).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "文章不存在")
		return
	}
	// 草稿仅管理员（携带有效 token）可见
	if article.Status != model.StatusPublished && optionalUserID(c) == 0 {
		response.Fail(c, http.StatusNotFound, "文章不存在")
		return
	}

	model.DB.Model(&model.Article{}).Where("id = ?", article.ID).
		UpdateColumn("views", gorm.Expr("views + 1"))

	type neighbor struct {
		ID    uint   `json:"id"`
		Title string `json:"title"`
	}
	var prev, next *neighbor
	var p model.Article
	if err := model.DB.Where("status = ? AND type = ? AND id < ?",
		model.StatusPublished, article.Type, article.ID).
		Order("id DESC").First(&p).Error; err == nil {
		prev = &neighbor{ID: p.ID, Title: p.Title}
	}
	var n model.Article
	if err := model.DB.Where("status = ? AND type = ? AND id > ?",
		model.StatusPublished, article.Type, article.ID).
		Order("id ASC").First(&n).Error; err == nil {
		next = &neighbor{ID: n.ID, Title: n.Title}
	}

	response.OK(c, gin.H{"article": article, "prev": prev, "next": next})
}

type archiveItem struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Type      int       `json:"type"`
	Language  string    `json:"language"`
	CreatedAt time.Time `json:"createdAt"`
}

type archiveGroup struct {
	Month string        `json:"month"`
	Items []archiveItem `json:"items"`
}

func Archives(c *gin.Context) {
	var items []archiveItem
	if err := model.DB.Model(&model.Article{}).
		Select("id, title, type, language, created_at").
		Where("status = ?", model.StatusPublished).
		Order("created_at DESC").Find(&items).Error; err != nil {
		response.ServerError(c, err)
		return
	}

	var groups []archiveGroup
	index := map[string]int{}
	for _, it := range items {
		month := it.CreatedAt.Format("2006-01")
		if _, ok := index[month]; !ok {
			index[month] = len(groups)
			groups = append(groups, archiveGroup{Month: month})
		}
		groups[index[month]].Items = append(groups[index[month]].Items, it)
	}
	response.OK(c, groups)
}

// ---------- 管理接口 ----------

type articleForm struct {
	Title       string `json:"title" binding:"required"`
	Summary     string `json:"summary"`
	Cover       string `json:"cover"`
	Content     string `json:"content" binding:"required"`
	Translation string `json:"translation"`
	CategoryID  uint   `json:"categoryId"`
	Type        int    `json:"type"`
	Language    string `json:"language"`
	Status      int    `json:"status"`
	TagIDs      []uint `json:"tagIds"`
}

func AdminListArticles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	query := model.DB.Model(&model.Article{}).Preload("Category").Preload("Tags")
	if t := c.Query("type"); t != "" {
		query = query.Where("type = ?", t)
	}
	if kw := strings.TrimSpace(c.Query("keyword")); kw != "" {
		like := "%" + kw + "%"
		query = query.Where("title LIKE ?", like)
	}

	var total int64
	query.Count(&total)

	var articles []model.Article
	if err := query.Select("id, title, summary, cover, category_id, type, language, status, views, created_at, updated_at").
		Order("id DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&articles).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.Page(c, articles, total)
}

func CreateArticle(c *gin.Context) {
	var form articleForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "标题和正文不能为空")
		return
	}
	if form.Language == "" {
		form.Language = "zh"
	}
	article := model.Article{
		Title:       form.Title,
		Summary:     form.Summary,
		Cover:       form.Cover,
		Content:     form.Content,
		Translation: form.Translation,
		CategoryID:  form.CategoryID,
		Type:        form.Type,
		Language:    form.Language,
		Status:      form.Status,
	}
	if err := model.DB.Create(&article).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	applyTags(&article, form.TagIDs)

	model.DB.Preload("Category").Preload("Tags").First(&article, article.ID)
	response.OK(c, article)
}

func UpdateArticle(c *gin.Context) {
	var article model.Article
	if err := model.DB.First(&article, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "文章不存在")
		return
	}
	var form articleForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "标题和正文不能为空")
		return
	}
	if form.Language == "" {
		form.Language = "zh"
	}
	article.Title = form.Title
	article.Summary = form.Summary
	article.Cover = form.Cover
	article.Content = form.Content
	article.Translation = form.Translation
	article.CategoryID = form.CategoryID
	article.Type = form.Type
	article.Language = form.Language
	article.Status = form.Status
	if err := model.DB.Save(&article).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	applyTags(&article, form.TagIDs)

	model.DB.Preload("Category").Preload("Tags").First(&article, article.ID)
	response.OK(c, article)
}

func DeleteArticle(c *gin.Context) {
	var article model.Article
	if err := model.DB.First(&article, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "文章不存在")
		return
	}
	model.DB.Model(&article).Association("Tags").Clear()
	model.DB.Where("article_id = ?", article.ID).Delete(&model.Annotation{})
	model.DB.Where("article_id = ?", article.ID).Delete(&model.Comment{})
	model.DB.Delete(&article)
	response.OK(c, nil)
}

func applyTags(article *model.Article, tagIDs []uint) {
	var tags []model.Tag
	if len(tagIDs) > 0 {
		model.DB.Find(&tags, tagIDs)
	}
	model.DB.Model(article).Association("Tags").Replace(&tags)
}
