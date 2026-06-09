package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"stdoutcms/internal/store"
)

// Version is set at build time via ldflags.
var Version = "dev"

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ---- Meta ----

type MetaResponse struct {
	App       string `json:"app"`
	Version   string `json:"version"`
	Uptime    string `json:"uptime"`
	Posts     int    `json:"posts"`
	Projects  int    `json:"projects"`
	GoVersion string `json:"goVersion"`
	AI        bool   `json:"ai"` // LLM features enabled
}

var startTime = time.Now()

func Meta(pg *store.Postgres, llmEnabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		posts, _ := pg.CountPosts()
		projects, _ := pg.CountProjects()
		c.JSON(http.StatusOK, MetaResponse{
			App:       "STDOUT_CMS_ELF",
			Version:   Version,
			Uptime:    time.Since(startTime).Truncate(time.Second).String(),
			Posts:     posts,
			Projects:  projects,
			GoVersion: runtime.Version(),
			AI:        llmEnabled,
		})
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
