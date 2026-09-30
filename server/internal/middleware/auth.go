package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"blogwitgoofy/server/internal/config"
	"blogwitgoofy/server/pkg/response"
	"blogwitgoofy/server/pkg/utils"
)

// JWTAuth 校验 Authorization: Bearer <token>
func JWTAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			response.Unauthorized(c, "未登录或 token 缺失")
			c.Abort()
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims, err := utils.ParseToken(tokenStr, cfg.JWT.Secret)
		if err != nil {
			response.Unauthorized(c, "token 无效或已过期")
			c.Abort()
			return
		}
		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
