package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"stdoutcms/internal/store"
)

// ---- Public Posts ----

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

// ---- Admin Posts ----

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
		if po.Tags == nil {
			po.Tags = []string{}
		}
		if po.Slug == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "slug is required"})
			return
		}
		if !slugRegex.MatchString(po.Slug) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "slug must be lowercase alphanumeric with hyphens only"})
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
		if po.Tags == nil {
			po.Tags = []string{}
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
