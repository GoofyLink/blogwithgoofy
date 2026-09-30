package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"blogwitgoofy/server/internal/model"
	"blogwitgoofy/server/pkg/response"
)

// ---------- 公开接口：章节笔记（读） ----------

func ListChapterNotes(c *gin.Context) {
	var notes []model.ChapterNote
	if err := model.DB.Where("chapter_id = ?", c.Param("id")).
		Order("paragraph_index ASC, id ASC").Find(&notes).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, notes)
}

// ---------- 管理接口：章节笔记（增删改） ----------

type chapterNoteForm struct {
	ChapterID      uint   `json:"chapterId" binding:"required"`
	ParagraphIndex int    `json:"paragraphIndex"`
	Quote          string `json:"quote" binding:"max=1000"`
	Note           string `json:"note" binding:"required,max=10000"`
}

func CreateChapterNote(c *gin.Context) {
	var form chapterNoteForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "笔记内容不能为空")
		return
	}
	var chapter model.Chapter
	if err := model.DB.First(&chapter, form.ChapterID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "章节不存在")
		return
	}
	note := model.ChapterNote{
		ChapterID:      form.ChapterID,
		ParagraphIndex: form.ParagraphIndex,
		Quote:          form.Quote,
		Note:           form.Note,
	}
	if err := model.DB.Create(&note).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, note)
}

func UpdateChapterNote(c *gin.Context) {
	var note model.ChapterNote
	if err := model.DB.First(&note, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "笔记不存在")
		return
	}
	var form struct {
		ParagraphIndex int    `json:"paragraphIndex"`
		Quote          string `json:"quote" binding:"max=1000"`
		Note           string `json:"note" binding:"required,max=10000"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "笔记内容不能为空")
		return
	}
	note.ParagraphIndex = form.ParagraphIndex
	note.Quote = form.Quote
	note.Note = form.Note
	if err := model.DB.Save(&note).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, note)
}

func DeleteChapterNote(c *gin.Context) {
	var note model.ChapterNote
	if err := model.DB.First(&note, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "笔记不存在")
		return
	}
	model.DB.Delete(&note)
	response.OK(c, nil)
}
