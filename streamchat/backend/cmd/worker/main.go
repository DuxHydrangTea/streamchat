package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"streamchat/backend/internal/config"
	"streamchat/backend/internal/database"
	"streamchat/backend/internal/media"
	"streamchat/backend/internal/storage"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("configuration invalid", "error", err)
		os.Exit(1)
	}
	db, err := database.Open(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Error("database unavailable", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	runner := media.ProcessRunner{Limits: media.Limits{MaxSizeBytes: cfg.MediaMaxSizeMB * 1024 * 1024, MaxDuration: time.Duration(cfg.MediaMaxDurationSeconds) * time.Second, MaxHeight: cfg.MediaMaxHeight, Timeout: time.Duration(cfg.MediaDownloadTimeoutSeconds) * time.Second, TempRoot: cfg.MediaTempDir, YTDLP: cfg.YTDLPBinary, FFprobe: cfg.FFprobeBinary, Proxy: cfg.MediaEgressProxy}}
	logger.Info("media worker listening", "concurrency", cfg.MediaWorkerConcurrency, "storage", "local")
	if err := media.RunServer(media.Worker{Jobs: media.JobStore{DB: db}, Runner: runner, Store: storage.Local{Root: cfg.LocalStorageRoot}, Logger: logger}, cfg.RedisAddr, cfg.MediaWorkerConcurrency); err != nil {
		logger.Error("worker stopped", "error", err)
		os.Exit(1)
	}
}
