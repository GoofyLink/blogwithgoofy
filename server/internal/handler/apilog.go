package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"blogwitgoofy/server/internal/model"
	"blogwitgoofy/server/pkg/response"
)

// AdminListApiLogs 接口日志列表（分页 + 筛选：路径/IP/方法）
func AdminListApiLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	query := model.DB.Model(&model.ApiLog{})
	if kw := strings.TrimSpace(c.Query("keyword")); kw != "" {
		like := "%" + kw + "%"
		query = query.Where("path LIKE ? OR ip LIKE ?", like, like)
	}
	if m := c.Query("method"); m != "" {
		query = query.Where("method = ?", m)
	}
	if st := c.Query("status"); st != "" {
		query = query.Where("status = ?", st)
	}

	var total int64
	query.Count(&total)

	var logs []model.ApiLog
	if err := query.Order("id DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&logs).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.Page(c, logs, total)
}

// ClearApiLogs 清空全部接口日志
func ClearApiLogs(c *gin.Context) {
	model.DB.Where("1 = 1").Delete(&model.ApiLog{})
	response.OK(c, nil)
}

// ApiLogStats 日志概览：总调用数、今日调用数、独立 IP 数
func ApiLogStats(c *gin.Context) {
	var total, today, ips int64
	model.DB.Model(&model.ApiLog{}).Count(&total)
	model.DB.Model(&model.ApiLog{}).
		Where("created_at >= ?", todayStart()).Count(&today)
	model.DB.Model(&model.ApiLog{}).
		Select("COUNT(DISTINCT ip)").Scan(&ips)
	response.OK(c, gin.H{"total": total, "today": today, "ips": ips})
}

func todayStart() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}
