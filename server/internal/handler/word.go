package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"

	"blogwitgoofy/server/internal/model"
	"blogwitgoofy/server/pkg/response"
)

// ---------- 公开接口：生词本（读） ----------

func ListWords(c *gin.Context) {
	query := model.DB.Model(&model.Word{})
	if lang := c.Query("language"); lang != "" {
		query = query.Where("language = ?", lang)
	}
	if kw := strings.TrimSpace(c.Query("keyword")); kw != "" {
		query = query.Where("word LIKE ? OR translation LIKE ?", "%"+kw+"%", "%"+kw+"%")
	}
	var words []model.Word
	if err := query.Order("id DESC").Limit(2000).Find(&words).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, words)
}

// ---------- 管理接口：生词本（收藏/删除） ----------

type wordForm struct {
	Word        string `json:"word" binding:"required,max=191"`
	Language    string `json:"language" binding:"required,oneof=en de"`
	Translation string `json:"translation" binding:"max=500"`
	SourceText  string `json:"sourceText" binding:"max=500"`
}

// AddWord 收藏生词；重复收藏时更新释义
func AddWord(c *gin.Context) {
	var form wordForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "单词无效")
		return
	}
	word := model.Word{
		Word:        strings.TrimSpace(form.Word),
		Language:    form.Language,
		Translation: form.Translation,
		SourceText:  form.SourceText,
	}
	if err := model.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "word"}, {Name: "language"}},
		DoUpdates: clause.AssignmentColumns([]string{"translation"}), // 重复收藏仅更新释义，保留原语境
	}).Create(&word).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	// 回读拿到最终记录（冲突更新时主键不变）
	var saved model.Word
	model.DB.Where("word = ? AND language = ?", word.Word, word.Language).First(&saved)
	response.OK(c, saved)
}

func DeleteWord(c *gin.Context) {
	var word model.Word
	if err := model.DB.First(&word, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "生词不存在")
		return
	}
	model.DB.Delete(&word)
	response.OK(c, nil)
}
