package handler

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"blogwitgoofy/server/internal/model"
	"blogwitgoofy/server/pkg/response"
)

// ---------- 公开接口：动漫排行榜 ----------

func ListAnimes(c *gin.Context) {
	var animes []model.Anime
	if err := model.DB.Where("status = ?", model.StatusPublished).
		Order("rank ASC, id ASC").Find(&animes).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, animes)
}

// ---------- 管理接口 ----------

type animeForm struct {
	Title          string   `json:"title" binding:"required,max=191"`
	Cover          string   `json:"cover" binding:"max=500"`
	Category       string   `json:"category" binding:"max=100"`
	Region         string   `json:"region" binding:"max=100"`
	Description    string   `json:"description" binding:"max=1000"`
	Rank           int      `json:"rank" binding:"min=0"`
	Status         int      `json:"status" binding:"oneof=0 1"`
	WatchStatus    string   `json:"watchStatus" binding:"omitempty,oneof=planned watching completed paused dropped"`
	AirStatus      string   `json:"airStatus" binding:"omitempty,oneof=upcoming airing finished"`
	Watched        int      `json:"watched" binding:"min=0"`
	TotalEpisodes  *int     `json:"totalEpisodes" binding:"omitempty,min=1"`
	Rating         *float64 `json:"rating" binding:"omitempty,min=0,max=10"`
	Year           int      `json:"year" binding:"min=0,max=2100"`
	Source         string   `json:"source" binding:"max=100"`
	Review         string   `json:"review" binding:"max=300"`
	Content        string   `json:"content" binding:"max=200000"`
	Spoiler        bool     `json:"spoiler"`
	Recommended    bool     `json:"recommended"`
	RecommendOrder int      `json:"recommendOrder" binding:"min=0"`
	WatchURL       string   `json:"watchUrl" binding:"max=1000"`
	WatchPlatform  string   `json:"watchPlatform" binding:"max=100"`
}

func (f animeForm) valid() bool {
	if strings.TrimSpace(f.Title) == "" || (f.TotalEpisodes != nil && f.Watched > *f.TotalEpisodes) {
		return false
	}
	if f.WatchURL != "" {
		u, err := url.Parse(f.WatchURL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
			return false
		}
	}
	return true
}
func (f animeForm) apply(a *model.Anime) {
	now := time.Now()
	if a.Watched != f.Watched || a.WatchStatus != f.WatchStatus {
		a.ProgressUpdatedAt = &now
	}
	if a.Content != f.Content {
		a.ReviewUpdatedAt = &now
	}
	a.Title = strings.TrimSpace(f.Title)
	a.Cover = f.Cover
	a.Category = strings.TrimSpace(f.Category)
	a.Region = f.Region
	a.Description = f.Description
	a.Rank = f.Rank
	a.Status = f.Status
	a.WatchStatus = f.WatchStatus
	a.AirStatus = f.AirStatus
	a.Watched = f.Watched
	a.TotalEpisodes = f.TotalEpisodes
	a.Rating = f.Rating
	a.Year = f.Year
	a.Source = f.Source
	a.Review = f.Review
	a.Content = f.Content
	a.Spoiler = f.Spoiler
	a.Recommended = f.Recommended
	a.RecommendOrder = f.RecommendOrder
	a.WatchURL = f.WatchURL
	a.WatchPlatform = f.WatchPlatform
}

func AdminListAnimes(c *gin.Context) {
	var animes []model.Anime
	if err := model.DB.Order("rank ASC, id ASC").Find(&animes).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, animes)
}

func CreateAnime(c *gin.Context) {
	var form animeForm
	if err := c.ShouldBindJSON(&form); err != nil || !form.valid() {
		response.BadRequest(c, "请检查标题、评分、集数和观看链接，已看集数不能超过总集数")
		return
	}
	var anime model.Anime
	form.apply(&anime)
	if err := model.DB.Create(&anime).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, anime)
}

func UpdateAnime(c *gin.Context) {
	var anime model.Anime
	if err := model.DB.First(&anime, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "动漫不存在")
		return
	}
	var form animeForm
	if err := c.ShouldBindJSON(&form); err != nil || !form.valid() {
		response.BadRequest(c, "请检查标题、评分、集数和观看链接，已看集数不能超过总集数")
		return
	}
	form.apply(&anime)
	if err := model.DB.Save(&anime).Error; err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, anime)
}

func DeleteAnime(c *gin.Context) {
	var anime model.Anime
	if err := model.DB.First(&anime, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "动漫不存在")
		return
	}
	model.DB.Delete(&anime)
	response.OK(c, nil)
}

func GetAnime(c *gin.Context) {
	var anime model.Anime
	if err := model.DB.Where("status = ?", model.StatusPublished).First(&anime, c.Param("id")).Error; err != nil {
		response.Fail(c, http.StatusNotFound, "作品不存在或已下架")
		return
	}
	response.OK(c, anime)
}

// Atomic increment prevents lost updates when several requests arrive together.
func AdvanceAnime(c *gin.Context) {
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		var anime model.Anime
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&anime, c.Param("id")).Error; err != nil {
			return err
		}
		if anime.TotalEpisodes != nil && anime.Watched >= *anime.TotalEpisodes {
			return errors.New("已看完全部集数")
		}
		anime.Watched++
		anime.WatchStatus = "watching"
		if anime.TotalEpisodes != nil && anime.Watched == *anime.TotalEpisodes {
			anime.WatchStatus = "completed"
		}
		now := time.Now()
		anime.ProgressUpdatedAt = &now
		return tx.Save(&anime).Error
	})
	if err != nil {
		response.BadRequest(c, "无法更新进度：作品不存在、已看完或更新失败")
		return
	}
	response.OK(c, nil)
}
