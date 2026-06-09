package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"stdoutcms/internal/config"
	"stdoutcms/internal/store"
)

func Auth(rd *store.Redis, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// login 接口本身不需要鉴权
		if c.Request.URL.Path == "/api/v1/admin/login" && c.Request.Method == "POST" {
			c.Next()
			return
		}

		var token string
		// 优先从 Authorization header 读
		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			token = strings.TrimPrefix(auth, "Bearer ")
		} else {
			// fallback 到 cookie
			token, _ = c.Cookie("blog_session_token")
		}

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		valid, err := rd.ValidateSession(c.Request.Context(), token)
		if err != nil || !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session expired"})
			return
		}

		c.Set("session_token", token)
		c.Next()
	}
}
