package middleware

import "github.com/gin-gonic/gin"

// CORS returns a middleware that sets CORS headers for same-origin requests.
// It does NOT reflect arbitrary Origin headers — cross-origin requests with
// credentials are deliberately blocked. In production, Nginx serves both
// frontend and backend on the same origin. In dev, use Vite's proxy.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Only set Allow-Origin for same-origin requests (Origin header absent)
		// or for explicitly allowed origins (e.g. local dev server).
		if origin == "" {
			// Same-origin — safe
		} else if origin == "http://localhost:5173" {
			// Vite dev server — allow with credentials
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		// All other origins: no Allow-Origin, no credentials

		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
