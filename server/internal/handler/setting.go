package handler

import (
	"github.com/gin-gonic/gin"

	"blogwitgoofy/server/internal/model"
	"blogwitgoofy/server/pkg/response"
)

// 默认值：未配置时前台使用
var defaultSettings = map[string]string{
	"nickname":     "小天",
	"bio":          "我在这里记录编程、英语与德语学习，也写一些故事和生活随笔。目前正在用 Vue 3 和 Go 持续完善这个博客。",
	"avatar":       "",
	"github":       "https://github.com/",
	"email":        "",
	"now_dev":      "小天 Goofy Blog · Vue 3 + Go",
	"now_learning": "英语 · 德语",
	"now_reading":  "《三体》 刘慈欣",
	"now_update":   "博客上线了多主题切换与小说专栏",
}

// GetSettings 公开读取站点配置（合并默认值）
func GetSettings(c *gin.Context) {
	var rows []model.Setting
	if err := model.DB.Find(&rows).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	merged := make(map[string]string, len(defaultSettings))
	for k, v := range defaultSettings {
		merged[k] = v
	}
	for _, r := range rows {
		if r.Value != "" {
			merged[r.Key] = r.Value
		}
	}
	response.OK(c, merged)
}

// SaveSettings 批量保存站点配置（upsert）
func SaveSettings(c *gin.Context) {
	var form map[string]string
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "配置格式无效")
		return
	}
	for k, v := range form {
		if len(k) > 100 {
			continue
		}
		setting := model.Setting{Key: k, Value: v}
		model.DB.Where("`key` = ?", k).Assign(setting).FirstOrCreate(&setting)
	}
	response.OK(c, nil)
}
