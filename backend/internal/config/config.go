package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppPort         string
	AppEnv          string
	DatabaseURL     string
	RedisURL        string
	SessionSecret   string
	SessionMaxAge   int // seconds
	Storage         StorageConfig
	LLM             LLMConfig
}

type LLMConfig struct {
	Endpoint     string
	APIKey       string
	Model        string
	DailyLimit   int
	SystemPrompt string // empty → use built-in default
}

// IsEnabled returns true when all required LLM fields are configured.
// If the user doesn't set LLM_ENDPOINT / LLM_API_KEY / LLM_MODEL,
// AI features are silently disabled across both backend and frontend.
func (c LLMConfig) IsEnabled() bool {
	return c.Endpoint != "" && c.APIKey != "" && c.Model != ""
}

type StorageConfig struct {
	Provider   string
	Endpoint   string
	Bucket     string
	Region     string
	AccessKey  string
	SecretKey  string
	CDNBaseURL string
	UseSSL     bool
}

func Load() *Config {
	return &Config{
		AppPort:       getEnv("APP_PORT", "8080"),
		AppEnv:        getEnv("APP_ENV", "development"),
		DatabaseURL:   mustEnv("DATABASE_URL"),
		RedisURL:      mustEnv("REDIS_URL"),
		SessionSecret: mustEnv("SESSION_SECRET"),
		SessionMaxAge: getEnvInt("SESSION_MAX_AGE", 86400*7), // 7 days
		Storage: StorageConfig{
			Provider:   getEnv("STORAGE_PROVIDER", "s3"),
			Endpoint:   getEnv("STORAGE_ENDPOINT", "localhost:9000"),
			Bucket:     getEnv("STORAGE_BUCKET", "blog-assets"),
			Region:     getEnv("STORAGE_REGION", ""),
			AccessKey:  mustEnv("STORAGE_ACCESS_KEY"),
			SecretKey:  mustEnv("STORAGE_SECRET_KEY"),
			CDNBaseURL: getEnv("STORAGE_CDN_BASE_URL", "http://localhost:9000/blog-assets"),
			UseSSL:     getEnvBool("STORAGE_USE_SSL", false),
		},
		LLM: LLMConfig{
			Endpoint:     getEnv("LLM_ENDPOINT", ""),
			APIKey:       getEnv("LLM_API_KEY", ""),
			Model:        getEnv("LLM_MODEL", ""),
			DailyLimit:   getEnvInt("LLM_DAILY_LIMIT", 50),
			SystemPrompt: getEnv("LLM_SYSTEM_PROMPT", ""),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("missing required env: " + key)
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
