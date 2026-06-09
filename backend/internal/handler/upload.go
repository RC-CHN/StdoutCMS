package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"stdoutcms/internal/config"
	"stdoutcms/internal/storage"
)

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
