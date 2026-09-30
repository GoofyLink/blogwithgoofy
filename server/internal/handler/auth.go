package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"blogwitgoofy/server/internal/config"
	"blogwitgoofy/server/internal/model"
	"blogwitgoofy/server/pkg/response"
	"blogwitgoofy/server/pkg/utils"
)

// Cfg 由 router.Setup 注入
var Cfg *config.Config

func Init(cfg *config.Config) { Cfg = cfg }

type loginForm struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var form loginForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "请输入用户名和密码")
		return
	}
	var user model.User
	if err := model.DB.Where("username = ?", form.Username).First(&user).Error; err != nil {
		response.Unauthorized(c, "用户名或密码错误")
		return
	}
	if !utils.CheckPassword(user.PasswordHash, form.Password) {
		response.Unauthorized(c, "用户名或密码错误")
		return
	}
	token, err := utils.GenerateToken(user.ID, user.Username, Cfg.JWT.Secret, Cfg.JWT.ExpireDays)
	if err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, gin.H{"token": token, "user": user})
}

type passwordForm struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

func ChangePassword(c *gin.Context) {
	var form passwordForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "新密码长度至少 6 位")
		return
	}
	userID := c.GetUint("userId")
	var user model.User
	if err := model.DB.First(&user, userID).Error; err != nil {
		response.Unauthorized(c, "用户不存在")
		return
	}
	if !utils.CheckPassword(user.PasswordHash, form.OldPassword) {
		response.BadRequest(c, "原密码错误")
		return
	}
	hash, err := utils.HashPassword(form.NewPassword)
	if err != nil {
		response.ServerError(c, err)
		return
	}
	model.DB.Model(&user).Update("password_hash", hash)
	response.OK(c, nil)
}

// optionalUserID 尝试解析 token，失败返回 0（用于公开接口识别管理员身份）
func optionalUserID(c *gin.Context) uint {
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return 0
	}
	claims, err := utils.ParseToken(strings.TrimPrefix(auth, "Bearer "), Cfg.JWT.Secret)
	if err != nil {
		return 0
	}
	return claims.UserID
}
