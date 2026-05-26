package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"stdoutcms/internal/config"
	"stdoutcms/internal/store"
)

// ---- types ----

type ChatRequest struct {
	Q   string `json:"q"`
	Ctx string `json:"ctx"`
	Sid string `json:"sid"`
}

type ChatResponse struct {
	A   string `json:"a"`
	Sid string `json:"sid"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type openAIResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// ---- Chat (public) ----

func Chat(pg *store.Postgres, rd *store.Redis, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. parse body
		var req ChatRequest
		if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Q) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "q is required"})
			return
		}
		if len(req.Q) > 500 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "q too long, max 500 chars"})
			return
		}

		// 2. rate limit
		if _, ok := rd.ChatDailyQuota(cfg.LLM.DailyLimit); !ok {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "daily quota exceeded"})
			return
		}

		// 3. resolve ctx → article content
		var articleContent string
		if req.Ctx != "" {
			post, err := pg.GetPostBySlug(req.Ctx)
			if err == nil && post != nil {
				articleContent = post.Content
			}
		}

		// delegate to shared logic
		chat, err := doChat(rd, cfg, req.Sid, req.Q, articleContent)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, chat)
	}
}

// ---- AdminChat (no rate limit, behind auth) ----

func AdminChat(pg *store.Postgres, rd *store.Redis, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ChatRequest
		if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Q) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "q is required"})
			return
		}
		if len(req.Q) > 500 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "q too long, max 500 chars"})
			return
		}

		var articleContent string
		if req.Ctx != "" {
			post, err := pg.GetPostBySlug(req.Ctx)
			if err == nil && post != nil {
				articleContent = post.Content
			}
		}

		chat, err := doChat(rd, cfg, req.Sid, req.Q, articleContent)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, chat)
	}
}

// ---- core ----

func doChat(rd *store.Redis, cfg *config.Config, sid, q, articleContent string) (*ChatResponse, error) {
	// ensure sid
	if sid == "" {
		sid = "chat_" + uuid.New().String()
	}

	// load history
	history, err := rd.ChatGetHistory(sid)
	if err != nil {
		return nil, fmt.Errorf("redis read: %w", err)
	}

	// build messages array
	messages := make([]chatMessage, 0, len(history)+3)

	// system prompt
	system := chatMessage{Role: "system", Content: buildSystemPrompt(articleContent)}
	messages = append(messages, system)

	// history
	for _, raw := range history {
		var msg chatMessage
		if err := json.Unmarshal([]byte(raw), &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	// user message
	userMsg := chatMessage{Role: "user", Content: q}
	messages = append(messages, userMsg)

	// call LLM
	answer, err := callLLM(cfg, messages)
	if err != nil {
		return nil, fmt.Errorf("llm: %w", err)
	}

	// persist
	userJSON, _ := json.Marshal(userMsg)
	assistantJSON, _ := json.Marshal(chatMessage{Role: "assistant", Content: answer})
	_ = rd.ChatPushMessage(sid, string(userJSON), time.Hour)
	_ = rd.ChatPushMessage(sid, string(assistantJSON), time.Hour)

	return &ChatResponse{A: answer, Sid: sid}, nil
}

// ---- helpers ----

func buildSystemPrompt(articleContent string) string {
	base := `你是一个运行在博客系统中的终端风格 AI 助手，代号 STDOUT_CMS_ELF。
你的回复应该简洁、准确，风格与博客的终端/brutalist 设计语言一致。`

	if articleContent != "" {
		// truncate very long articles to avoid blowing context
		truncated := articleContent
		if len(truncated) > 4000 {
			truncated = truncated[:4000] + "..."
		}
		base += fmt.Sprintf("\n\n用户正在阅读以下文章，请基于文章内容回答问题：\n\n```\n%s\n```", truncated)
	}

	return base
}

func callLLM(cfg *config.Config, messages []chatMessage) (string, error) {
	body := openAIRequest{
		Model:    cfg.LLM.Model,
		Messages: messages,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	url := strings.TrimRight(cfg.LLM.Endpoint, "/") + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.LLM.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result openAIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse llm response: %w (body=%s)", err, string(respBody))
	}

	if result.Error != nil {
		return "", fmt.Errorf("llm api error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("llm returned no choices")
	}

	return result.Choices[0].Message.Content, nil
}
