package router

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDevelopmentOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(developmentCORS())
	for _, tc := range []struct {
		origin string
		want   int
	}{
		{"http://localhost:5180", 204}, {"http://127.0.0.1:5180", 204},
		{"http://localhost:5173", 204}, {"https://untrusted.example", 403},
	} {
		t.Run(tc.origin, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/login", nil)
			req.Header.Set("Origin", tc.origin)
			req.Header.Set("Access-Control-Request-Method", "POST")
			req.Header.Set("Access-Control-Request-Headers", "content-type")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.want {
				t.Fatalf("got %d, want %d", w.Code, tc.want)
			}
			if tc.want == 204 && w.Header().Get("Access-Control-Allow-Origin") != tc.origin {
				t.Fatal("missing allowed origin")
			}
		})
	}
}
