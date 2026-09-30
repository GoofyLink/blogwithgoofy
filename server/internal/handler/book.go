package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"blogwitgoofy/server/internal/model"
	"blogwitgoofy/server/pkg/response"
)

// ---------- 公开接口 ----------

// ListBooks 书架列表（含章节数），type 过滤：0 学习书房 / 1 小说专栏
func ListBooks(c *gin.Context) {
	var books []model.Book
	query := model.DB.Model(&model.Book{}).
		Select("books.*, COUNT(chapters.id) AS chapter_count").
		Joins("LEFT JOIN chapters ON chapters.book_id = books.id AND chapters.status = ?", model.StatusPublished).
		Group("books.id").
		Where("books.status = ?", model.StatusPublished)
	if t := c.Query("type"); t != "" {
		query = query.Where("books.type = ?", t)
	}
	if lang := c.Query("language"); lang != "" {
		query = query.Where("books.language = ?", lang)
	}
	if err := query.Order("books.sort ASC, books.id ASC").Scan(&books).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, books)
}

// GetBook 书籍详情 + 章节目录（不含正文）
func GetBook(c *gin.Context) {
	var book model.Book
	if err := model.DB.First(&book, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "书籍不存在")
		return
	}

	type chapterItem struct {
		ID        uint   `json:"id"`
		Title     string `json:"title"`
		Sort      int    `json:"sort"`
		CreatedAt string `json:"createdAt"`
	}
	var chapters []chapterItem
	if err := model.DB.Model(&model.Chapter{}).
		Select("id, title, sort, created_at").
		Where("book_id = ? AND status = ?", book.ID, model.StatusPublished).
		Order("sort ASC, id ASC").Scan(&chapters).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, gin.H{"book": book, "chapters": chapters})
}

// GetChapter 章节正文（+ 所属书 + 上一章/下一章）
func GetChapter(c *gin.Context) {
	var chapter model.Chapter
	if err := model.DB.Where("status = ?", model.StatusPublished).
		First(&chapter, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "章节不存在")
		return
	}

	var book model.Book
	if err := model.DB.First(&book, chapter.BookID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "所属书籍不存在")
		return
	}

	type neighbor struct {
		ID    uint   `json:"id"`
		Title string `json:"title"`
	}
	var prev, next *neighbor
	var p model.Chapter
	if err := model.DB.Where("book_id = ? AND status = ? AND (sort < ? OR (sort = ? AND id < ?))",
		chapter.BookID, model.StatusPublished, chapter.Sort, chapter.Sort, chapter.ID).
		Order("sort DESC, id DESC").First(&p).Error; err == nil {
		prev = &neighbor{ID: p.ID, Title: p.Title}
	}
	var n model.Chapter
	if err := model.DB.Where("book_id = ? AND status = ? AND (sort > ? OR (sort = ? AND id > ?))",
		chapter.BookID, model.StatusPublished, chapter.Sort, chapter.Sort, chapter.ID).
		Order("sort ASC, id ASC").First(&n).Error; err == nil {
		next = &neighbor{ID: n.ID, Title: n.Title}
	}

	response.OK(c, gin.H{"chapter": chapter, "book": book, "prev": prev, "next": next})
}

// ---------- 管理接口：书籍 ----------

type bookForm struct {
	Title       string `json:"title" binding:"required,max=191"`
	Subtitle    string `json:"subtitle" binding:"max=191"`
	Author      string `json:"author" binding:"max=191"`
	Language    string `json:"language" binding:"required,oneof=en de zh"`
	Cover       string `json:"cover" binding:"max=500"`
	Description string `json:"description" binding:"max=2000"`
	Type        int    `json:"type"` // 0 学习书房 / 1 小说专栏
	Status      int    `json:"status"`
	Sort        int    `json:"sort"`
}

func AdminListBooks(c *gin.Context) {
	var books []model.Book
	query := model.DB.Model(&model.Book{}).
		Select("books.*, COUNT(chapters.id) AS chapter_count").
		Joins("LEFT JOIN chapters ON chapters.book_id = books.id").
		Group("books.id")
	if err := query.Order("books.sort ASC, books.id ASC").Scan(&books).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, books)
}

func CreateBook(c *gin.Context) {
	var form bookForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "请填写书名，语言仅支持 en/de/zh")
		return
	}
	book := model.Book{
		Title: form.Title, Subtitle: form.Subtitle, Author: form.Author,
		Language: form.Language, Cover: form.Cover, Description: form.Description,
		Type: form.Type, Status: form.Status, Sort: form.Sort,
	}
	if err := model.DB.Create(&book).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, book)
}

func UpdateBook(c *gin.Context) {
	var book model.Book
	if err := model.DB.First(&book, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "书籍不存在")
		return
	}
	var form bookForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "请填写书名，语言仅支持 en/de/zh")
		return
	}
	book.Title = form.Title
	book.Subtitle = form.Subtitle
	book.Author = form.Author
	book.Language = form.Language
	book.Cover = form.Cover
	book.Description = form.Description
	book.Type = form.Type
	book.Status = form.Status
	book.Sort = form.Sort
	if err := model.DB.Save(&book).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, book)
}

func DeleteBook(c *gin.Context) {
	var book model.Book
	if err := model.DB.First(&book, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "书籍不存在")
		return
	}
	model.DB.Where("chapter_id IN (SELECT id FROM chapters WHERE book_id = ?)", book.ID).
		Delete(&model.ChapterNote{})
	model.DB.Where("book_id = ?", book.ID).Delete(&model.Chapter{})
	model.DB.Delete(&book)
	response.OK(c, nil)
}

// ---------- 管理接口：章节 ----------

type chapterForm struct {
	BookID      uint   `json:"bookId" binding:"required"`
	Title       string `json:"title" binding:"required,max=191"`
	Content     string `json:"content" binding:"required"`
	Translation string `json:"translation"`
	Sort        int    `json:"sort"`
	Status      int    `json:"status"`
}

// AdminListChapters 某本书的全部章节（管理用，含正文概要）
func AdminListChapters(c *gin.Context) {
	bookID := c.Param("id")
	var chapters []model.Chapter
	if err := model.DB.Where("book_id = ?", bookID).
		Order("sort ASC, id ASC").Find(&chapters).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, chapters)
}

func CreateChapter(c *gin.Context) {
	var form chapterForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "请填写章节标题和内容")
		return
	}
	var book model.Book
	if err := model.DB.First(&book, form.BookID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "所属书籍不存在")
		return
	}
	chapter := model.Chapter{
		BookID: form.BookID, Title: form.Title, Content: form.Content,
		Translation: form.Translation, Sort: form.Sort, Status: form.Status,
	}
	if err := model.DB.Create(&chapter).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, chapter)
}

func UpdateChapter(c *gin.Context) {
	var chapter model.Chapter
	if err := model.DB.First(&chapter, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "章节不存在")
		return
	}
	var form struct {
		Title       string `json:"title" binding:"required,max=191"`
		Content     string `json:"content" binding:"required"`
		Translation string `json:"translation"`
		Sort        int    `json:"sort"`
		Status      int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "请填写章节标题和内容")
		return
	}
	chapter.Title = form.Title
	chapter.Content = form.Content
	chapter.Translation = form.Translation
	chapter.Sort = form.Sort
	chapter.Status = form.Status
	if err := model.DB.Save(&chapter).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, chapter)
}

func DeleteChapter(c *gin.Context) {
	var chapter model.Chapter
	if err := model.DB.First(&chapter, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "章节不存在")
		return
	}
	model.DB.Where("chapter_id = ?", chapter.ID).Delete(&model.ChapterNote{})
	model.DB.Delete(&chapter)
	response.OK(c, nil)
}

// ---------- 批量导入章节 ----------

type batchChapterItem struct {
	Title       string `json:"title" binding:"required,max=191"`
	Content     string `json:"content"` // 允许为空：仅标题快速建章，内容后续补充
	Translation string `json:"translation"`
}

type batchChapterForm struct {
	BookID    uint                `json:"bookId" binding:"required"`
	StartSort int                  `json:"startSort"`
	Chapters  []batchChapterItem  `json:"chapters" binding:"required,min=1,max=500"`
}

// BatchCreateChapters 一次导入整本书的章节（前端解析好章节数组后提交）
func BatchCreateChapters(c *gin.Context) {
	var form batchChapterForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "章节列表无效：每章需有标题和内容，单次最多 500 章")
		return
	}
	var book model.Book
	if err := model.DB.First(&book, form.BookID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "所属书籍不存在")
		return
	}

	sort := form.StartSort
	if sort <= 0 {
		var maxSort int
		model.DB.Model(&model.Chapter{}).Where("book_id = ?", form.BookID).
			Select("COALESCE(MAX(sort), 0)").Scan(&maxSort)
		sort = maxSort
	}

	err := model.DB.Transaction(func(tx *gorm.DB) error {
		for i, item := range form.Chapters {
			chapter := model.Chapter{
				BookID:      form.BookID,
				Title:       item.Title,
				Content:     item.Content,
				Translation: item.Translation,
				Sort:        sort + i + 1,
				Status:      model.StatusPublished,
			}
			if err := tx.Create(&chapter).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, gin.H{"created": len(form.Chapters)})
}
