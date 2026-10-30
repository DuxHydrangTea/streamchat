package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Environment                 string
	APIPort                     string
	DatabaseURL                 string
	RedisAddr                   string
	JWTSecret                   string
	CookieSecure                bool
	FrontendOrigin              string
	MediaMaxSizeMB              int64
	MediaMaxDurationSeconds     int64
	MediaMaxHeight              int
	MediaDownloadTimeoutSeconds int
	MediaWorkerConcurrency      int
	YTDLPBinary                 string
	FFmpegBinary                string
	FFprobeBinary               string
	MediaEgressProxy            string
	MediaTempDir                string
	LocalStorageRoot            string
}

func Load() (Config, error) {
	cfg := Config{
		Environment:                 env("APP_ENV", "development"),
		APIPort:                     env("API_PORT", "8080"),
		DatabaseURL:                 os.Getenv("DATABASE_URL"),
		RedisAddr:                   env("REDIS_ADDR", "localhost:6379"),
		JWTSecret:                   os.Getenv("JWT_SECRET"),
		FrontendOrigin:              env("FRONTEND_ORIGIN", "http://localhost:5173"),
		MediaMaxSizeMB:              int64Env("MEDIA_MAX_SIZE_MB", 200),
		MediaMaxDurationSeconds:     int64Env("MEDIA_MAX_DURATION_SECONDS", 600),
		MediaMaxHeight:              intEnv("MEDIA_MAX_HEIGHT", 720),
		MediaDownloadTimeoutSeconds: intEnv("MEDIA_DOWNLOAD_TIMEOUT_SECONDS", 300),
		MediaWorkerConcurrency:      intEnv("MEDIA_WORKER_CONCURRENCY", 2),
		YTDLPBinary:                 env("YTDLP_BINARY", "yt-dlp"),
		FFmpegBinary:                env("FFMPEG_BINARY", "ffmpeg"),
		FFprobeBinary:               env("FFPROBE_BINARY", "ffprobe"),
		MediaEgressProxy:            os.Getenv("MEDIA_EGRESS_PROXY"),
		MediaTempDir:                env("MEDIA_TEMP_DIR", "/tmp/streamchat-media"),
		LocalStorageRoot:            env("LOCAL_STORAGE_ROOT", "./data"),
	}
	secure, err := strconv.ParseBool(env("COOKIE_SECURE", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("COOKIE_SECURE: %w", err)
	}
	cfg.CookieSecure = secure
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET must contain at least 32 characters")
	}
	if cfg.MediaMaxSizeMB < 1 || cfg.MediaMaxDurationSeconds < 1 || cfg.MediaMaxHeight < 1 || cfg.MediaDownloadTimeoutSeconds < 1 || cfg.MediaWorkerConcurrency < 1 {
		return Config{}, errors.New("media limits and concurrency must be positive")
	}
	if cfg.APIPort == "" {
		return Config{}, errors.New("API_PORT cannot be empty")
	}
	return cfg, nil
}

func intEnv(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}
func int64Env(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
