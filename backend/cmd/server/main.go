package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"stdoutcms/internal/config"
	"stdoutcms/internal/handler"
	"stdoutcms/internal/middleware"
	"stdoutcms/internal/storage"
	"stdoutcms/internal/store"
)

// defaultPath tries a few locations for the migration file.
func defaultMigrationPath() string {
	candidates := []string{
		"migrations/001_init.sql",
		"../migrations/001_init.sql",
		"../../migrations/001_init.sql",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			abs, _ := filepath.Abs(p)
			return abs
		}
	}
	return "migrations/001_init.sql"
}

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	pg, err := store.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect postgres", "error", err)
		os.Exit(1)
	}
	defer pg.Close()

	// auto-migration
	migrationPath := defaultMigrationPath()
	sql, err := os.ReadFile(migrationPath)
	if err != nil {
		logger.Error("failed to read migration file", "error", err)
		os.Exit(1)
	}
	if err := pg.ExecMigration(string(sql)); err != nil {
		logger.Error("failed to run migration", "error", err)
		os.Exit(1)
	}
	logger.Info("migration complete")

	rd, err := store.NewRedis(cfg.RedisURL)
	if err != nil {
		logger.Error("failed to connect redis", "error", err)
		os.Exit(1)
	}
	defer rd.Close()

	s3, err := storage.NewS3(cfg.Storage)
	if err != nil {
		logger.Error("failed to connect s3", "error", err)
		os.Exit(1)
	}

	// seed admin on startup
	adminUser := os.Getenv("ADMIN_USERNAME")
	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminUser != "" && adminPass != "" {
		if err := handler.SeedAdmin(pg, adminUser, adminPass); err != nil {
			logger.Error("failed to seed admin", "error", err)
			os.Exit(1)
		}
		logger.Info("admin user seeded")
	}

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger(logger))
	r.Use(middleware.CORS())

	// health
	r.GET("/health", handler.Health)
	r.GET("/api/v1/meta", handler.Meta(pg, cfg.LLM.IsEnabled()))

	// public
	r.GET("/api/v1/posts", handler.ListPosts(pg, rd))
	r.GET("/api/v1/posts/:slug", handler.GetPost(pg, rd))
	r.GET("/api/v1/projects", handler.ListProjects(pg, rd))
	r.GET("/api/v1/projects/:id", handler.GetProject(pg))

	// AI chat (public) — only if LLM is configured
	if cfg.LLM.IsEnabled() {
		r.POST("/api/v1/ai/chat", handler.Chat(pg, rd, cfg))
	}

	// admin
	admin := r.Group("/api/v1/admin")
	admin.Use(middleware.Auth(rd, cfg))
	{
		admin.POST("/login", handler.Login(pg, rd, cfg))
		admin.DELETE("/logout", handler.Logout(rd, cfg))
		admin.GET("/posts", handler.ListPostsAdmin(pg))
		admin.GET("/posts/:slug", handler.GetPostAdmin(pg))
		admin.POST("/posts", handler.CreatePost(pg, rd))
		admin.PUT("/posts/:slug", handler.UpdatePost(pg, rd))
		admin.DELETE("/posts/:slug", handler.DeletePost(pg, rd))
		admin.POST("/upload", handler.UploadImage(s3, cfg))
		admin.GET("/projects", handler.ListProjectsAdmin(pg))
		admin.POST("/projects", handler.CreateProject(pg, rd))
		admin.PUT("/projects/:id", handler.UpdateProject(pg, rd))
		admin.DELETE("/projects/:id", handler.DeleteProject(pg, rd))

		// AI chat (admin) — only if LLM is configured
		if cfg.LLM.IsEnabled() {
			admin.POST("/ai/chat", handler.AdminChat(pg, rd, cfg))
		}
	}

	logger.Info("server starting", "port", cfg.AppPort)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
