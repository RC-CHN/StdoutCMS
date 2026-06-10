package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"stdoutcms/internal/config"
	"stdoutcms/internal/store"
)

func Login(pg *store.Postgres, rd *store.Redis, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		admin, err := pg.GetAdminByUsername(ctx, req.Username)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		if bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)) != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		token := generateToken()
		rd.SetSession(ctx, token, req.Username, cfg.SessionMaxAge)

		// Respect X-Forwarded-Proto from Nginx for cookie Secure flag.
		// When Nginx terminates TLS, it sets X-Forwarded-Proto=https.
		proto := c.GetHeader("X-Forwarded-Proto")
		secure := proto == "https"
		c.SetCookie("blog_session_token", token, cfg.SessionMaxAge, "/", "", secure, true)
		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}

func Logout(rd *store.Redis, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		token, _ := c.Get("session_token")
		if t, ok := token.(string); ok {
			rd.DeleteSession(ctx, t)
		}
		c.SetCookie("blog_session_token", "", -1, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
