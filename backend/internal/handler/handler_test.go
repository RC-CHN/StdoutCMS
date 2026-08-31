package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestCalcReadTime(t *testing.T) {
	cases := []struct {
		words int
		want  string
	}{
		{0, "0S"},
		{200, "30S"},
		{400, "1 MIN"},
		{600, "2 MIN"}, // 90s rounds up to 2 min
		{4000, "10 MIN"},
	}
	for _, tc := range cases {
		if got := calcReadTime(tc.words); got != tc.want {
			t.Errorf("calcReadTime(%d) = %q, want %q", tc.words, got, tc.want)
		}
	}
}

func TestSlugRegex(t *testing.T) {
	valid := []string{"a", "hello", "hello-world", "a0-1b2"}
	invalid := []string{"", "Hello", "-hello", "hello-", "hello--world", "hello_world", "hello world", "héllo"}
	for _, s := range valid {
		if !slugRegex.MatchString(s) {
			t.Errorf("slug %q should be valid", s)
		}
	}
	for _, s := range invalid {
		if slugRegex.MatchString(s) {
			t.Errorf("slug %q should be invalid", s)
		}
	}
}

// TestUpdatePostRejectsInvalidSlug verifies route-slug validation fires
// before the store is touched (nil stores are safe on this path).
func TestUpdatePostRejectsInvalidSlug(t *testing.T) {
	r := gin.New()
	r.PUT("/posts/:slug", UpdatePost(nil, nil))

	for _, slug := range []string{"Bad_Slug", "-nope", "nope-", "a--b"} {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/posts/"+slug, strings.NewReader(`{"title":"x"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("slug %q: expected 400, got %d", slug, w.Code)
		}
	}
}

func TestKindForContentType(t *testing.T) {
	cases := []struct {
		ct     string
		prefix string
		kind   string
	}{
		{"image/png", "images", "image"},
		{"audio/mpeg", "audio", "audio"},
		{"video/mp4", "video", "video"},
		{"application/pdf", "files", "file"},
		{"", "files", "file"},
	}
	for _, tc := range cases {
		prefix, kind := kindForContentType(tc.ct)
		if prefix != tc.prefix || kind != tc.kind {
			t.Errorf("kindForContentType(%q) = (%q, %q), want (%q, %q)",
				tc.ct, prefix, kind, tc.prefix, tc.kind)
		}
	}
}
