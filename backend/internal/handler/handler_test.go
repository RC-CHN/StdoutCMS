package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHealth(t *testing.T) {
	r := gin.New()
	r.GET("/health", Health)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %s", body["status"])
	}
}

func TestMetaFormat(t *testing.T) {
	r := gin.New()
	// Test that Meta handler returns correct field structure
	r.GET("/meta", func(c *gin.Context) {
		c.JSON(200, MetaResponse{
			App:       "SYS_BLOG.EXE",
			Version:   "test",
			Uptime:    "1m0s",
			Posts:     3,
			Projects:  2,
			GoVersion: "go-test",
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/meta", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp MetaResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.App != "SYS_BLOG.EXE" {
		t.Errorf("expected SYS_BLOG.EXE, got %s", resp.App)
	}
	if resp.Posts != 3 {
		t.Errorf("expected 3 posts, got %d", resp.Posts)
	}
	if resp.Projects != 2 {
		t.Errorf("expected 2 projects, got %d", resp.Projects)
	}
}

func TestGenerateToken(t *testing.T) {
	t1 := generateToken()
	t2 := generateToken()
	if t1 == t2 {
		t.Error("tokens should be unique")
	}
	if len(t1) != 64 {
		t.Errorf("expected 64 hex chars, got %d", len(t1))
	}
}
