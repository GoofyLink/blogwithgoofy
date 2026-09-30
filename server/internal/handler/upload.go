package handler

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"blogwitgoofy/server/pkg/response"
)

var allowedImageExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

// Upload 图片上传，返回 /uploads/xxx 可访问地址
func Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的图片")
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExt[ext] {
		response.BadRequest(c, "仅支持 jpg/png/gif/webp 格式图片")
		return
	}
	if file.Size > 5<<20 {
		response.BadRequest(c, "图片大小不能超过 5MB")
		return
	}

	name := fmt.Sprintf("%d%s", time.Now().UnixMilli(), ext)
	if err := c.SaveUploadedFile(file, filepath.Join("uploads", name)); err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, gin.H{"url": "/uploads/" + name})
}
