package media

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"mime"
	"os"
	"path/filepath"

	"streamchat/backend/internal/storage"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type Worker struct {
	Jobs   JobStore
	Runner Runner
	Store  storage.Store
	Logger *slog.Logger
}

func (w Worker) HandleDownload(ctx context.Context, task *asynq.Task) error {
	var job Job
	if err := json.Unmarshal(task.Payload(), &job); err != nil {
		return err
	}
	if _, err := uuid.Parse(job.ID); err != nil {
		return errors.New("invalid media job id")
	}
	if w.Store == nil {
		return errors.New("media storage is not configured")
	}
	if err := w.Jobs.SetStatus(ctx, job.ID, "fetching_metadata", "", ""); err != nil {
		return err
	}
	path, cleanup, err := w.Runner.Download(ctx, job.SourceURL)
	if err != nil {
		_ = w.Jobs.SetStatus(ctx, job.ID, "failed", "download_failed", "media download failed")
		return err
	}
	if cleanup != nil {
		defer cleanup()
	} else {
		defer os.Remove(path)
	}
	metadata, err := w.Runner.Probe(ctx, path)
	if err != nil {
		_ = w.Jobs.SetStatus(ctx, job.ID, "failed", "probe_failed", "media validation failed")
		return err
	}
	if err := w.Jobs.SetStatus(ctx, job.ID, "uploading", "", ""); err != nil {
		return err
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Size() < 1 || (w.RunnerLimitBytes() > 0 && info.Size() > w.RunnerLimitBytes()) {
		_ = w.Jobs.SetStatus(ctx, job.ID, "failed", "size_limit", "media exceeds configured size")
		return errors.New("media exceeds configured size")
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	key := "media/" + job.ID + filepath.Ext(path)
	contentType := mime.TypeByExtension(filepath.Ext(path))
	if contentType == "" {
		contentType = "video/mp4"
	}
	if err = w.Store.Put(ctx, key, file, contentType, info.Size()); err != nil {
		_ = file.Close()
		_ = w.Jobs.SetStatus(ctx, job.ID, "failed", "storage_failed", "media storage failed")
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := w.Jobs.Complete(ctx, job, key, contentType, info.Size(), metadata); err != nil {
		return err
	}
	return nil
}

func (w Worker) RunnerLimitBytes() int64 {
	if runner, ok := w.Runner.(ProcessRunner); ok {
		return runner.Limits.MaxSizeBytes
	}
	return 0
}

func RunServer(w Worker, redisAddr string, concurrency int) error {
	srv := asynq.NewServer(asynq.RedisClientOpt{Addr: redisAddr}, asynq.Config{Concurrency: concurrency})
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskDownloadVideo, w.HandleDownload)
	return srv.Run(mux)
}
