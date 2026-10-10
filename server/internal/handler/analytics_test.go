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
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func testAnalyticsID(n int) string { return fmt.Sprintf("ab128abc-0000-4000-8000-%012d", n) }
func testAnalyticsEvent(n int, now time.Time) analyticsEvent {
	return analyticsEvent{ID: testAnalyticsID(n), ViewID: testAnalyticsID(n), Visitor: testAnalyticsID(8001), Session: testAnalyticsID(9001), Path: "/", Kind: "page_view", StartedAt: now.Add(-100 * time.Second).UnixMilli()}
}
func TestAnalyticsValidation(t *testing.T) {
	now := time.Now()
	e := testAnalyticsEvent(1, now)
	if !validAnalyticsEvent(e, now) {
		t.Fatal("valid event rejected")
	}
	for _, mutate := range []func(*analyticsEvent){func(e *analyticsEvent) { e.Visitor = "email@example.com" }, func(e *analyticsEvent) { e.Path = "/?password=private" }, func(e *analyticsEvent) { e.Seconds = -1 }, func(e *analyticsEvent) { e.Seconds = 14401 }, func(e *analyticsEvent) { e.StartedAt = now.Add(-48 * time.Hour).UnixMilli() }, func(e *analyticsEvent) { e.Kind = "unknown" }, func(e *analyticsEvent) { e.Kind = "click"; e.Action = "unknown" }} {
		bad := e
		mutate(&bad)
		if validAnalyticsEvent(bad, now) {
			t.Fatal("accepted invalid event")
		}
	}
	if analyticsHost("javascript:alert(1)") != "" {
		t.Fatal("unsafe target")
	}
	if analyticsSource("https://www.google.com/search?q=private", "localhost:8080") != "搜索引擎：google.com" {
		t.Fatal("search attribution")
	}
	if analyticsSource("https://google.com.evil.example/path?private=1", "localhost:8080") != "google.com.evil.example" {
		t.Fatal("incorrect domain classification")
	}
	if analyticsHost("https://example.com/private?q=secret") != "example.com" {
		t.Fatal("query leaked")
	}
	if analyticsDevice("Mozilla iPhone Mobile") != "手机" || analyticsDevice("iPad") != "平板" {
		t.Fatal("device classification")
	}
}
func TestAnalyticsDateRange(t *testing.T) {
	now := time.Date(2026, 10, 10, 17, 0, 0, 0, time.UTC)
	a, b, err := analyticsDates("", "", now)
	if err != nil || a.Format("2006-01-02") != "2026-10-05" || b.Format("2006-01-02") != "2026-10-11" {
		t.Fatal("Beijing dates")
	}
	for _, pair := range [][2]string{{"2026-10-12", "2026-10-12"}, {"2025-01-01", "2026-10-01"}, {"2026-10-10", "2026-10-09"}, {"bad", "bad"}} {
		if _, _, err := analyticsDates(pair[0], pair[1], now); err == nil {
			t.Fatal("invalid range accepted")
		}
	}
}
func TestAnalyticsDatabase(t *testing.T) {
	dir := os.Getenv("ANALYTICS_TEST_CONFIG_DIR")
	if dir == "" {
		t.Skip("set ANALYTICS_TEST_CONFIG_DIR for local database integration")
	}
	viper.AddConfigPath(dir)
	cfg := config.Load()
	oldCfg := Cfg
	Cfg = cfg
	defer func() { Cfg = oldCfg }()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfg.MySQL.Username, cfg.MySQL.Password, cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal("database unavailable")
	}
	if err = db.AutoMigrate(&model.AnalyticsVisit{}, &model.AnalyticsFact{}, &model.AnalyticsReceipt{}, &model.AnalyticsState{}); err != nil {
		t.Fatal(err)
	}
	db.FirstOrCreate(&model.AnalyticsState{ID: 1})
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	original := model.DB
	model.DB = tx
	defer func() { tx.Rollback(); model.DB = original }()
	now := time.Now().In(analyticsZone)
	yesterday := now.AddDate(0, 0, -1)
	read := func(include bool) analyticsMetrics {
		gin.SetMode(gin.TestMode)
		r := gin.New()
		r.GET("/stats", AnalyticsDashboard)
		w := httptest.NewRecorder()
		url := fmt.Sprintf("/stats?start=%s&end=%s&module=home&includeAdmin=%t", yesterday.Format("2006-01-02"), now.Format("2006-01-02"), include)
		r.ServeHTTP(w, httptest.NewRequest("GET", url, nil))
		if w.Code != 200 {
			t.Fatalf("dashboard %s", w.Body.String())
		}
		var out struct {
			Data struct {
				Metrics analyticsMetrics `json:"metrics"`
			} `json:"data"`
		}
		if err = json.Unmarshal(w.Body.Bytes(), &out); err != nil {
			t.Fatal(err)
		}
		return out.Data.Metrics
	}
	before := read(false)
	beforeAll := read(true)
	write := func(e analyticsEvent, admin bool, at time.Time) {
		t.Helper()
		if err := tx.Transaction(func(q *gorm.DB) error { return ingestAnalytics(q, e, admin, "直接访问", "电脑", at) }); err != nil {
			t.Fatal(err)
		}
	}
	e := testAnalyticsEvent(1, yesterday)
	write(e, false, yesterday)
	write(e, false, yesterday)
	e.ID = testAnalyticsID(2)
	e.Kind = "page_engagement"
	e.Seconds = 30
	write(e, false, yesterday)
	e.ID = testAnalyticsID(3)
	e.Seconds = 20
	write(e, false, yesterday) // out-of-order cumulative report
	e.ID = testAnalyticsID(4)
	e.Seconds = 45
	write(e, false, yesterday)
	next := testAnalyticsEvent(5, now)
	write(next, false, now) // same visitor and session, another day
	other := testAnalyticsEvent(6, now)
	other.Visitor = testAnalyticsID(8002)
	other.Session = testAnalyticsID(9002)
	write(other, false, now)
	admin := testAnalyticsEvent(7, now)
	admin.Visitor = testAnalyticsID(8003)
	write(admin, true, now)
	click := next
	click.ID = testAnalyticsID(8)
	click.Kind = "click"
	click.Action = "outbound"
	click.Target = "https://example.com/path?secret=1"
	write(click, false, now)
	write(click, false, now)
	after := read(false)
	all := read(true)
	if after.PV-before.PV != 3 || after.UV-before.UV != 2 || after.Sessions-before.Sessions != 2 || after.Seconds-before.Seconds != 45 {
		t.Fatalf("dedupe/UV/duration mismatch before=%+v after=%+v", before, after)
	}
	if all.PV-beforeAll.PV != 4 || all.UV-beforeAll.UV != 3 {
		t.Fatal("admin filter failed")
	}
	var fact model.AnalyticsFact
	if err = tx.Where("visitor=? AND action='outbound'", analyticsHash(next.Visitor)).First(&fact).Error; err != nil {
		t.Fatal(err)
	}
	if fact.Count != 1 || fact.Target != "example.com" {
		t.Fatal("click dedupe/target sanitization failed")
	}
	if _, _, err = resolveAnalyticsPage(tx, "/admin/dashboard"); err == nil {
		t.Fatal("admin path accepted")
	}
	category := model.Category{Name: "__analytics_test__"}
	if err = tx.Create(&category).Error; err != nil {
		t.Fatal(err)
	}
	article := model.Article{Title: "__analytics_test__", Type: 1, Status: 1, CategoryID: category.ID}
	if err = tx.Create(&article).Error; err != nil {
		t.Fatal(err)
	}
	m, _, err := resolveAnalyticsPage(tx, fmt.Sprintf("/article/%d", article.ID))
	if err != nil || m != "learn" {
		t.Fatal("learning article attribution")
	}
	tx.Model(&article).Update("status", 0)
	if _, _, err = resolveAnalyticsPage(tx, fmt.Sprintf("/article/%d", article.ID)); err == nil {
		t.Fatal("draft counted")
	}
}
