package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"stdoutcms/internal/config"
	"stdoutcms/internal/storage"
)

// kindForContentType maps a MIME type to a storage prefix and a response kind tag.
// Falls back to "files"/"file" for unrecognised types.
func kindForContentType(ct string) (prefix string, kind string) {
	switch {
	case strings.HasPrefix(ct, "image/"):
		return "images", "image"
	case strings.HasPrefix(ct, "audio/"):
		return "audio", "audio"
	case strings.HasPrefix(ct, "video/"):
		return "video", "video"
	default:
		return "files", "file"
	}
}

func UploadFile(s3 *storage.S3, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing file"})
			return
		}
		defer file.Close()

		contentType := header.Header.Get("Content-Type")
		prefix, kind := kindForContentType(contentType)

		// size limits: video 100MB, others 10MB
		maxSize := int64(10 << 20)
		if kind == "video" {
			maxSize = 100 << 20
		}
		if header.Size > maxSize {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file too large"})
			return
		}

		key := storage.GenerateKey(prefix, uuid.New().String(), contentType)
		url, err := s3.Upload(c.Request.Context(), key, file, header.Size, contentType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"url": url, "kind": kind})
	}
}
