package media

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Limits struct {
	MaxSizeBytes int64
	MaxDuration  time.Duration
	MaxHeight    int
	Timeout      time.Duration
	TempRoot     string
	YTDLP        string
	FFprobe      string
	Proxy        string
}
type Runner interface {
	Download(context.Context, string) (string, func(), error)
	Probe(context.Context, string) (Metadata, error)
}
type Metadata struct {
	Title    string  `json:"title"`
	Duration float64 `json:"duration"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	Ext      string  `json:"ext"`
	Filesize int64   `json:"filesize"`
	URL      string  `json:"url"`
}
type ProcessRunner struct{ Limits Limits }
type commandRunner func(context.Context, string, ...string) ([]byte, error)

func (r ProcessRunner) run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return runCommand(ctx, name, args...)
}
func (r ProcessRunner) Download(ctx context.Context, source string) (string, func(), error) {
	return r.downloadWithCommand(ctx, r.run, source)
}
func runCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		message := stderr.String()
		if len(message) > 2048 {
			message = message[:2048]
		}
		return nil, fmt.Errorf("process failed: %s", strings.TrimSpace(message))
	}
	return stdout.Bytes(), nil
}
func (r ProcessRunner) downloadWithCommand(ctx context.Context, run commandRunner, source string) (string, func(), error) {
	l := r.Limits
	if l.Timeout <= 0 {
		l.Timeout = 5 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, l.Timeout)
	defer cancel()
	if l.TempRoot == "" {
		return "", nil, errors.New("media temp directory is not configured")
	}
	if err := os.MkdirAll(l.TempRoot, 0700); err != nil {
		return "", nil, err
	}
	dir, err := os.MkdirTemp(l.TempRoot, "streamchat-media-")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	output := filepath.Join(dir, "video.%(ext)s")
	if strings.TrimSpace(l.Proxy) == "" {
		cleanup()
		return "", nil, errors.New("media egress proxy is not configured")
	}
	if l.MaxSizeBytes <= 0 || l.MaxDuration <= 0 || l.MaxHeight <= 0 {
		cleanup()
		return "", nil, errors.New("media limits must be positive")
	}
	proxy := l.Proxy
	args := []string{"--ignore-config", "--no-plugin-dirs", "--no-playlist", "--no-warnings", "--max-filesize", fmt.Sprintf("%d", l.MaxSizeBytes), "--match-filter", fmt.Sprintf("duration <= %d", int(l.MaxDuration.Seconds())), "-f", fmt.Sprintf("best[height<=%d][ext=mp4]/best[height<=%d]", l.MaxHeight, l.MaxHeight), "--proxy", proxy, "-o", output, "--", source}
	if _, err := run(ctx, l.YTDLP, args...); err != nil {
		cleanup()
		return "", nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		cleanup()
		return "", nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			path := filepath.Join(dir, entry.Name())
			info, err := entry.Info()
			if err != nil {
				cleanup()
				return "", nil, err
			}
			if info.Size() > l.MaxSizeBytes {
				cleanup()
				return "", nil, errors.New("media exceeds configured size")
			}
			return path, cleanup, nil
		}
	}
	cleanup()
	return "", nil, errors.New("downloader produced no file")
}
func (r ProcessRunner) Probe(ctx context.Context, path string) (Metadata, error) {
	l := r.Limits
	if l.Timeout <= 0 {
		l.Timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, l.Timeout)
	defer cancel()
	out, err := r.run(ctx, l.FFprobe, "-v", "error", "-show_entries", "format=duration:format_tags=title:stream=width,height", "-of", "json", path)
	if err != nil {
		return Metadata{}, err
	}
	var doc struct {
		Format struct {
			Duration string `json:"duration"`
			Tags     struct {
				Title string `json:"title"`
			} `json:"tags"`
		} `json:"format"`
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}
	if err = json.Unmarshal(out, &doc); err != nil {
		return Metadata{}, err
	}
	duration, _ := strconv.ParseFloat(doc.Format.Duration, 64)
	m := Metadata{Title: doc.Format.Tags.Title, Duration: duration}
	if len(doc.Streams) > 0 {
		m.Width = doc.Streams[0].Width
		m.Height = doc.Streams[0].Height
	}
	if duration > l.MaxDuration.Seconds() || m.Height > l.MaxHeight {
		return Metadata{}, errors.New("media exceeds configured duration or height")
	}
	if info, err := os.Stat(path); err == nil {
		m.Filesize = info.Size()
		if l.MaxSizeBytes > 0 && m.Filesize > l.MaxSizeBytes {
			return Metadata{}, errors.New("media exceeds configured size")
		}
	}
	return m, nil
}
func CopyBounded(dst io.Writer, src io.Reader, max int64) (int64, error) {
	n, err := io.Copy(dst, io.LimitReader(src, max+1))
	if n > max {
		return n, errors.New("file exceeds configured size")
	}
	return n, err
}
