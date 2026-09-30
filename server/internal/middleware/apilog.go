package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"blogwitgoofy/server/internal/model"
)

// apiLogBufferSize 日志缓冲区大小；写满后丢弃新日志，不影响请求
const apiLogBufferSize = 512

// apiLogRetention 日志保留天数
const apiLogRetention = 7 * 24 * time.Hour

// 不记录的高频/自查询路径前缀
var apiLogSkipPrefixes = []string{"/api/v1/admin/apilogs"}

// ApiLog 记录 /api/v1 接口调用：方法、路径、IP、状态码、耗时（异步落库）
func ApiLog() gin.HandlerFunc {
	ch := make(chan model.ApiLog, apiLogBufferSize)

	// 单一消费者：批量写入 + 定时清理过期日志
	go func() {
		batch := make([]model.ApiLog, 0, 32)
		flush := func() {
			if len(batch) == 0 {
				return
			}
			model.DB.Create(&batch)
			batch = batch[:0]
		}
		ticker := time.NewTicker(2 * time.Second)
		cleanTicker := time.NewTicker(1 * time.Hour)
		// 启动时先清一次
		model.DB.Where("created_at < ?", time.Now().Add(-apiLogRetention)).
			Delete(&model.ApiLog{})
		for {
			select {
			case log := <-ch:
				batch = append(batch, log)
				if len(batch) >= 32 {
					flush()
				}
			case <-ticker.C:
				flush()
			case <-cleanTicker.C:
				model.DB.Where("created_at < ?", time.Now().Add(-apiLogRetention)).
					Delete(&model.ApiLog{})
			}
		}
	}()

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.Request.URL.Path
		for _, p := range apiLogSkipPrefixes {
			if strings.HasPrefix(path, p) {
				return
			}
		}

		entry := model.ApiLog{
			Method:     c.Request.Method,
			Path:       path,
			IP:         c.ClientIP(),
			Status:     c.Writer.Status(),
			DurationMs: int(time.Since(start).Milliseconds()),
			UserAgent:  strings.TrimSpace(c.Request.UserAgent()),
		}
		if len(entry.UserAgent) > 480 {
			entry.UserAgent = entry.UserAgent[:480]
		}
		select {
		case ch <- entry:
		default: // 缓冲满则丢弃
		}
	}
}
