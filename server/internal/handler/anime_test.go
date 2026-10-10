package handler

import (
	"blogwitgoofy/server/internal/config"
	"blogwitgoofy/server/internal/model"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestAnimeRejectsInvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/animes", CreateAnime)
	for _, body := range []string{
		`{"title":" "}`, `{"title":"Test","rating":11}`, `{"title":"Test","rating":-1}`,
		`{"title":"Test","watched":13,"totalEpisodes":12}`, `{"title":"Test","totalEpisodes":0}`,
		`{"title":"Test","watchUrl":"javascript:alert(1)"}`, `{"title":"Test","watchUrl":"https://"}`,
		`{"title":"Test","watchStatus":"bad"}`, `{"title":"Test","airStatus":"bad"}`,
		`{"title":"Test","status":2}`, `{"title":"Test","watched":-1}`,
	} {
		t.Run(body, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/animes", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected rejection, got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestAnimeReviewAndProgressTimestamps(t *testing.T) {
	before := time.Now().Add(-time.Hour)
	a := model.Anime{Title: "Test", Content: "Original", Watched: 2, WatchStatus: "watching", ReviewUpdatedAt: &before, ProgressUpdatedAt: &before}
	f := animeForm{Title: "Test", Content: "Original", Watched: 3, WatchStatus: "watching", Status: 1}
	f.apply(&a)
	if !a.ReviewUpdatedAt.Equal(before) {
		t.Fatal("progress update must not bump review date")
	}
	if !a.ProgressUpdatedAt.After(before) {
		t.Fatal("progress timestamp not updated")
	}
	progress := *a.ProgressUpdatedAt
	f.Content = "Updated review"
	f.apply(&a)
	if !a.ReviewUpdatedAt.After(before) {
		t.Fatal("review timestamp not updated")
	}
	if !a.ProgressUpdatedAt.Equal(progress) {
		t.Fatal("review must not bump progress date")
	}
}

func TestAnimeNullableRatingAndEpisodes(t *testing.T) {
	zero := 0.0
	total := 12
	a := model.Anime{}
	f := animeForm{Title: "Test", Rating: &zero, TotalEpisodes: &total, Status: 0}
	f.apply(&a)
	if a.Rating == nil || *a.Rating != 0 || a.Status != 0 {
		t.Fatal("zero rating and hidden status must be preserved")
	}
	f.Rating = nil
	f.TotalEpisodes = nil
	f.apply(&a)
	if a.Rating != nil || a.TotalEpisodes != nil {
		t.Fatal("clearing optional fields failed")
	}
}

// Opt-in integration check: schema additions persist, test rows are rolled back.
func TestAnimeDatabaseRoundTrip(t *testing.T) {
	dir := os.Getenv("ANIME_TEST_CONFIG_DIR")
	if dir == "" {
		t.Skip("set ANIME_TEST_CONFIG_DIR to run against a local development database")
	}
	viper.AddConfigPath(dir)
	cfg := config.Load()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfg.MySQL.Username, cfg.MySQL.Password, cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal("cannot connect to development database")
	}
	if err = db.AutoMigrate(&model.Anime{}); err != nil {
		t.Fatal(err)
	}
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	original := model.DB
	model.DB = tx
	defer func() { tx.Rollback(); model.DB = original }()
	r := gin.New()
	r.POST("/animes", CreateAnime)
	r.PUT("/animes/:id", UpdateAnime)
	r.GET("/animes/:id", GetAnime)
	r.POST("/animes/:id/progress", AdvanceAnime)
	call := func(method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}
	body := `{"title":"__anime_transaction_test__","status":0,"watched":1,"totalEpisodes":2,"rating":9.5,"content":"## Review","spoiler":true,"recommended":true}`
	w := call("POST", "/animes", body)
	if w.Code != 200 {
		t.Fatalf("create: %s", w.Body.String())
	}
	var result struct {
		Data model.Anime `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	id := result.Data.ID
	path := fmt.Sprintf("/animes/%d", id)
	if result.Data.Status != 0 {
		t.Fatal("hidden create became published")
	}
	if w = call("GET", path, ""); w.Code != 404 {
		t.Fatal("hidden anime exposed")
	}
	if w = call("POST", path+"/progress", ""); w.Code != 200 {
		t.Fatalf("advance: %s", w.Body.String())
	}
	var got model.Anime
	tx.First(&got, id)
	if got.Watched != 2 || got.WatchStatus != "completed" {
		t.Fatal("last episode did not complete anime")
	}
	if w = call("POST", path+"/progress", ""); w.Code != 400 {
		t.Fatal("progress exceeded total episodes")
	}
	update := `{"title":"__anime_transaction_test__","status":1,"watched":2,"totalEpisodes":null,"rating":null,"content":"## Updated","watchStatus":"watching","recommended":false,"spoiler":false}`
	if w = call("PUT", path, update); w.Code != 200 {
		t.Fatalf("update: %s", w.Body.String())
	}
	w = call("GET", path, "")
	var published struct {
		Data model.Anime `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &published)
	if w.Code != 200 || published.Data.Rating != nil || published.Data.TotalEpisodes != nil || published.Data.Content != "## Updated" || published.Data.Recommended || published.Data.Spoiler {
		t.Fatal("published round trip or clearing fields failed")
	}
	if w = call("POST", path+"/progress", ""); w.Code != 200 {
		t.Fatal("unknown total must allow advancing")
	}
}
