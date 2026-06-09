package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"stdoutcms/internal/store"
)

// ---- Public About ----

func GetAbout(pg *store.Postgres, rd *store.Redis) gin.HandlerFunc {
	return func(c *gin.Context) {
		cacheKey := "about:page"
		if cached, ok := rd.GetCache(cacheKey); ok {
			c.Data(http.StatusOK, "application/json", cached)
			return
		}

		a, err := pg.GetAbout()
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "about not found"})
			return
		}
		body, _ := json.Marshal(a)
		rd.SetCache(cacheKey, body, 1*time.Hour)
		c.JSON(http.StatusOK, a)
	}
}

// ---- Admin About ----

func GetAboutAdmin(pg *store.Postgres) gin.HandlerFunc {
	return func(c *gin.Context) {
		a, err := pg.GetAbout()
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "about not found"})
			return
		}
		c.JSON(http.StatusOK, a)
	}
}

func UpdateAbout(pg *store.Postgres, rd *store.Redis) gin.HandlerFunc {
	return func(c *gin.Context) {
		var a store.About
		if err := c.ShouldBindJSON(&a); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := pg.UpdateAbout(a.Title, a.Content); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		rd.InvalidateAbout()
		c.JSON(http.StatusOK, gin.H{"status": "saved"})
	}
}
