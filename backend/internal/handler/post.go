package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"stdoutcms/internal/store"
)

// ---- Helpers ----

func calcReadTime(wordCount int) string {
	const charsPerMin = 400
	secs := int(math.Round(float64(wordCount) / float64(charsPerMin) * 60))
	if secs < 60 {
		return fmt.Sprintf("%dS", secs)
	}
	return fmt.Sprintf("%d MIN", int(math.Round(float64(secs)/60)))
}

// ---- Public Posts ----

func ListPosts(pg *store.Postgres, rd *store.Redis) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
		if page < 1 {
			page = 1
		}
		if size < 1 || size > 50 {
			size = 10
		}

		cacheKey := fmt.Sprintf("posts:list:%d:%d", page, size)
		if cached, ok := rd.GetCache(ctx, cacheKey); ok {
			c.Data(http.StatusOK, "application/json", cached)
			return
		}

		posts, total, err := pg.ListPosts(ctx, page, size)
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
		rd.SetCache(ctx, cacheKey, body, 5*time.Minute)
		c.JSON(http.StatusOK, resp)
	}
}

func GetPost(pg *store.Postgres, rd *store.Redis) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		slug := c.Param("slug")

		cacheKey := "posts:slug:" + slug
		if cached, ok := rd.GetCache(ctx, cacheKey); ok {
			c.Data(http.StatusOK, "application/json", cached)
			return
		}

		post, err := pg.GetPostBySlug(ctx, slug)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		body, _ := json.Marshal(post)
		rd.SetCache(ctx, cacheKey, body, 30*time.Minute)
		c.JSON(http.StatusOK, post)
	}
}

// ---- Admin Posts ----

func GetPostAdmin(pg *store.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		slug := c.Param("slug")
		post, err := pg.GetPostBySlugAdmin(ctx, slug)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		c.JSON(http.StatusOK, post)
	}
}

func ListPostsAdmin(pg *store.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
		if page < 1 {
			page = 1
		}
		if size < 1 || size > 50 {
			size = 10
		}

		posts, total, err := pg.ListPostsAdmin(ctx, page, size)
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
		ctx := c.Request.Context()
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
		if po.ReadTime == "" {
			po.ReadTime = calcReadTime(po.WordCount)
		}
		if err := pg.CreatePost(ctx, &po); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		rd.InvalidatePost(ctx, po.Slug)
		c.JSON(http.StatusCreated, gin.H{"slug": po.Slug})
	}
}

func UpdatePost(pg *store.Postgres, rd *store.Redis) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		slug := c.Param("slug")
		if !slugRegex.MatchString(slug) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "slug must be lowercase alphanumeric with hyphens only"})
			return
		}
		var po store.Post
		if err := c.ShouldBindJSON(&po); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if po.Tags == nil {
			po.Tags = []string{}
		}
		if po.ReadTime == "" {
			po.ReadTime = calcReadTime(po.WordCount)
		}
		if err := pg.UpdatePost(ctx, slug, &po); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		rd.InvalidatePost(ctx, slug)
		c.JSON(http.StatusOK, gin.H{"slug": slug, "status": "updated"})
	}
}

func DeletePost(pg *store.Postgres, rd *store.Redis) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		slug := c.Param("slug")
		if err := pg.DeletePost(ctx, slug); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		rd.InvalidatePost(ctx, slug)
		c.JSON(http.StatusOK, gin.H{"slug": slug, "status": "deleted"})
	}
}
