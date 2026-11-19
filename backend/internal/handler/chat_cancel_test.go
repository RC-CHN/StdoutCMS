package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"stdoutcms/internal/config"
)

func TestCancelActiveModelRequest(t *testing.T) {
	for _, mode := range []string{"chat", "metadata"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.Copy(io.Discard, r.Body)
				// Disconnect the caller while the model is still preparing its reply.
				cancel()
				select {
				case <-r.Context().Done():
				case <-time.After(3 * time.Second):
					t.Error("model request remained active after the caller disconnected")
				}
			}))
			defer upstream.Close()
			cfg := &config.Config{LLM: config.LLMConfig{Endpoint: upstream.URL}}
			var err error
			if mode == "chat" {
				c, _ := gin.CreateTestContext(httptest.NewRecorder())
				_, err = callLLMStream(ctx, cfg, nil, "test", c.Writer)
			} else {
				_, err = callLLMNonStream(ctx, cfg, nil)
			}
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("model request returned %v, want context.Canceled", err)
			}
		})
	}
}
