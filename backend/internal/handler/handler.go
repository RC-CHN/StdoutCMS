package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"stdoutcms/internal/config"
	"stdoutcms/internal/storage"
	"stdoutcms/internal/store"
)

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ---- Public ----

func ListPosts(pg *store.Postgres, rd *store.Redis) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
		if page < 1 {
			page = 1
		}
		if size < 1 || size > 50 {
			size = 10
		}

		cacheKey := fmt.Sprintf("posts:list:%d:%d", page, size)
		if cached, ok := rd.GetCache(cacheKey); ok {
			c.Data(http.StatusOK, "application/json", cached)
			return
		}

		posts, total, err := pg.ListPosts(page, size)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		resp := gin.H{
			"posts":    posts,
			"total":    total,
			"page":     page,
			"pageSize": size,
		}
		body, _ := json.Marshal(resp)
		rd.SetCache(cacheKey, body, 5*time.Minute)
		c.JSON(http.StatusOK, resp)
	}
}

func GetPost(pg *store.Postgres, rd *store.Redis) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")

		cacheKey := "posts:slug:" + slug
		if cached, ok := rd.GetCache(cacheKey); ok {
			c.Data(http.StatusOK, "application/json", cached)
			return
		}

		post, err := pg.GetPostBySlug(slug)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}
		body, _ := json.Marshal(post)
		rd.SetCache(cacheKey, body, 30*time.Minute)
		c.JSON(http.StatusOK, post)
	}
}

func GetPostAdmin(pg *store.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		post, err := pg.GetPostBySlugAdmin(slug)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}
		c.JSON(http.StatusOK, post)
	}
}

func ListProjects(pg *store.Postgres, rd *store.Redis) gin.HandlerFunc {
	return func(c *gin.Context) {
		cacheKey := "projects:list"
		if cached, ok := rd.GetCache(cacheKey); ok {
			c.Data(http.StatusOK, "application/json", cached)
			return
		}

		projects, err := pg.ListProjects()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		resp := gin.H{"projects": projects}
		body, _ := json.Marshal(resp)
		rd.SetCache(cacheKey, body, 10*time.Minute)
		c.JSON(http.StatusOK, resp)
	}
}

func GetProject(pg *store.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		project, err := pg.GetProject(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		c.JSON(http.StatusOK, project)
	}
}

// ---- Admin Auth ----

func Login(pg *store.Postgres, rd *store.Redis, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		admin, err := pg.GetAdminByUsername(req.Username)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		if bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)) != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		token := generateToken()
		rd.SetSession(token, req.Username, cfg.SessionMaxAge)

		secure := cfg.AppEnv == "production"
		c.SetCookie("blog_session_token", token, cfg.SessionMaxAge, "/", "", secure, true)
		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}

func Logout(rd *store.Redis, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := c.Get("session_token")
		if t, ok := token.(string); ok {
			rd.DeleteSession(t)
		}
		c.SetCookie("blog_session_token", "", -1, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

// ---- Admin Posts ----

func ListPostsAdmin(pg *store.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
		if page < 1 {
			page = 1
		}
		if size < 1 || size > 50 {
			size = 10
		}

		posts, total, err := pg.ListPostsAdmin(page, size)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"posts":    posts,
			"total":    total,
			"page":     page,
			"pageSize": size,
		})
	}
}

func CreatePost(pg *store.Postgres, rd *store.Redis) gin.HandlerFunc {
	return func(c *gin.Context) {
		var po store.Post
		if err := c.ShouldBindJSON(&po); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if po.Slug == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "slug is required"})
			return
		}
		if err := pg.CreatePost(&po); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		rd.InvalidatePost(po.Slug)
		c.JSON(http.StatusCreated, gin.H{"slug": po.Slug})
	}
}

func UpdatePost(pg *store.Postgres, rd *store.Redis) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		var po store.Post
		if err := c.ShouldBindJSON(&po); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := pg.UpdatePost(slug, &po); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		rd.InvalidatePost(slug)
		c.JSON(http.StatusOK, gin.H{"slug": slug, "status": "updated"})
	}
}

func DeletePost(pg *store.Postgres, rd *store.Redis) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		if err := pg.DeletePost(slug); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		rd.InvalidatePost(slug)
		c.JSON(http.StatusOK, gin.H{"slug": slug, "status": "deleted"})
	}
}

// ---- Admin Projects ----

func ListProjectsAdmin(pg *store.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		projects, err := pg.ListProjectsAdmin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"projects": projects})
	}
}

func CreateProject(pg *store.Postgres, rd *store.Redis) gin.HandlerFunc {
	return func(c *gin.Context) {
		var pr store.Project
		if err := c.ShouldBindJSON(&pr); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := pg.CreateProject(&pr); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		rd.InvalidateProjects()
		c.JSON(http.StatusCreated, gin.H{"status": "created"})
	}
}

func UpdateProject(pg *store.Postgres, rd *store.Redis) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		var pr store.Project
		if err := c.ShouldBindJSON(&pr); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := pg.UpdateProject(id, &pr); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		rd.InvalidateProjects()
		c.JSON(http.StatusOK, gin.H{"status": "updated"})
	}
}

func DeleteProject(pg *store.Postgres, rd *store.Redis) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		if err := pg.DeleteProject(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		rd.InvalidateProjects()
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	}
}

// ---- Admin Upload ----

func UploadImage(s3 *storage.S3, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing file"})
			return
		}
		defer file.Close()

		// 限制文件类型
		contentType := header.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "only images allowed"})
			return
		}

		// 限制大小 (10MB)
		const maxSize = 10 << 20
		if header.Size > maxSize {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 10MB)"})
			return
		}

		key := storage.GenerateKey(uuid.New().String(), contentType)
		url, err := s3.Upload(c.Request.Context(), key, file, header.Size, contentType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"url": url})
	}
}

// ---- Helpers ----

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// SeedAdmin creates the default admin user if none exists.
// Called once at startup from main.
func SeedAdmin(pg *store.Postgres, username, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("ADMIN_USERNAME and ADMIN_PASSWORD must be set")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return pg.SeedAdmin(username, string(hash))
}
