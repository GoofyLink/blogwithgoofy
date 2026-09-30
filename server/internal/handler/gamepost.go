package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"blogwitgoofy/server/internal/model"
	"blogwitgoofy/server/pkg/response"
)

// ---------- 公开接口 ----------

// GetGame 游戏详情
func GetGame(c *gin.Context) {
	var game model.Game
	if err := model.DB.Where("status = ?", model.StatusPublished).
		First(&game, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "游戏不存在")
		return
	}
	response.OK(c, game)
}

// ListGamePosts 某游戏的攻略资讯列表（type 筛选）
func ListGamePosts(c *gin.Context) {
	q := model.DB.Where("game_id = ? AND status = ?", c.Param("id"), model.StatusPublished)
	if t := c.Query("type"); t != "" {
		q = q.Where("type = ?", t)
	}
	var posts []model.GamePost
	if err := q.Order("id DESC").Find(&posts).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, posts)
}

// GetGamePost 帖子详情（浏览量 +1）
func GetGamePost(c *gin.Context) {
	var post model.GamePost
	if err := model.DB.Where("status = ?", model.StatusPublished).
		First(&post, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "帖子不存在")
		return
	}
	model.DB.Model(&model.GamePost{}).Where("id = ?", post.ID).
		UpdateColumn("views", gorm.Expr("views + 1"))
	var game model.Game
	model.DB.First(&game, post.GameID)
	response.OK(c, gin.H{"post": post, "game": game})
}

// ---------- 评论（公开读写） ----------

func ListGameComments(c *gin.Context) {
	var comments []model.GameComment
	if err := model.DB.Where("post_id = ?", c.Param("id")).
		Order("id ASC").Find(&comments).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, comments)
}

func CreateGameComment(c *gin.Context) {
	postID, _ := strconv.Atoi(c.Param("id"))
	var post model.GamePost
	if err := model.DB.First(&post, postID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "帖子不存在")
		return
	}
	var form struct {
		Nickname string `json:"nickname" binding:"required,max=100"`
		Email    string `json:"email" binding:"omitempty,email,max=191"`
		Content  string `json:"content" binding:"required,max=2000"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "请填写昵称和评论内容（邮箱格式需正确）")
		return
	}
	comment := model.GameComment{
		PostID: uint(postID), GameID: post.GameID,
		Nickname: strings.TrimSpace(form.Nickname),
		Email: form.Email, Content: form.Content,
	}
	if err := model.DB.Create(&comment).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, comment)
}

// ---------- 管理接口：帖子 ----------

type gamePostForm struct {
	GameID  uint   `json:"gameId" binding:"required"`
	Title   string `json:"title" binding:"required,max=191"`
	Type    string `json:"type" binding:"max=100"`
	Summary string `json:"summary" binding:"max=1000"`
	Content string `json:"content" binding:"required"`
	Status  int    `json:"status"`
}

func AdminListGamePosts(c *gin.Context) {
	var posts []model.GamePost
	q := model.DB.Order("id DESC")
	if gid := c.Query("gameId"); gid != "" {
		q = q.Where("game_id = ?", gid)
	}
	if err := q.Find(&posts).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, posts)
}

func CreateGamePost(c *gin.Context) {
	var form gamePostForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "请填写标题和内容")
		return
	}
	var game model.Game
	if err := model.DB.First(&game, form.GameID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "所属游戏不存在")
		return
	}
	if form.Type == "" {
		form.Type = "攻略"
	}
	post := model.GamePost{
		GameID: form.GameID, Title: form.Title, Type: form.Type,
		Summary: form.Summary, Content: form.Content, Status: form.Status,
	}
	if err := model.DB.Create(&post).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, post)
}

func UpdateGamePost(c *gin.Context) {
	var post model.GamePost
	if err := model.DB.First(&post, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "帖子不存在")
		return
	}
	var form gamePostForm
	if err := c.ShouldBindJSON(&form); err != nil {
		response.BadRequest(c, "请填写标题和内容")
		return
	}
	post.GameID = form.GameID
	post.Title = form.Title
	post.Type = form.Type
	post.Summary = form.Summary
	post.Content = form.Content
	post.Status = form.Status
	if err := model.DB.Save(&post).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, post)
}

func DeleteGamePost(c *gin.Context) {
	var post model.GamePost
	if err := model.DB.First(&post, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "帖子不存在")
		return
	}
	model.DB.Where("post_id = ?", post.ID).Delete(&model.GameComment{})
	model.DB.Delete(&post)
	response.OK(c, nil)
}

// ---------- 管理接口：评论 ----------

func AdminListGameComments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	query := model.DB.Model(&model.GameComment{}).
		Select("game_comments.*, game_posts.title AS post_title, games.title AS game_title").
		Joins("LEFT JOIN game_posts ON game_posts.id = game_comments.post_id").
		Joins("LEFT JOIN games ON games.id = game_comments.game_id")
	if pid := c.Query("postId"); pid != "" {
		query = query.Where("game_comments.post_id = ?", pid)
	}

	var total int64
	query.Count(&total)
	var comments []model.GameComment
	if err := query.Order("game_comments.id DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&comments).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.Page(c, comments, total)
}

func DeleteGameComment(c *gin.Context) {
	var comment model.GameComment
	if err := model.DB.First(&comment, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "评论不存在")
		return
	}
	model.DB.Delete(&comment)
	response.OK(c, nil)
}
