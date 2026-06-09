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

// ---- Public Projects ----

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

func GetProject(pg *store.Postgres, rd *store.Redis) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		cacheKey := fmt.Sprintf("projects:id:%d", id)
		if cached, ok := rd.GetCache(cacheKey); ok {
			c.Data(http.StatusOK, "application/json", cached)
			return
		}

		project, err := pg.GetProject(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}
		body, _ := json.Marshal(project)
		rd.SetCache(cacheKey, body, 10*time.Minute)
		c.JSON(http.StatusOK, project)
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
