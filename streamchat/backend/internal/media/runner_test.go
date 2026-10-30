package media

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestCopyBounded(t *testing.T) {
	var out strings.Builder
	if _, err := CopyBounded(&out, strings.NewReader("12345"), 4); err == nil {
		t.Fatal("expected size error")
	}
}

func TestProcessRunnerRequiresEgressProxy(t *testing.T) {
	runner := ProcessRunner{Limits: Limits{MaxSizeBytes: 200, MaxDuration: 10 * time.Second, MaxHeight: 720, TempRoot: t.TempDir()}}
	if _, _, err := runner.Download(context.Background(), "https://www.youtube.com/watch?v=x"); err == nil {
		t.Fatal("expected missing proxy error")
	}
}

func TestProcessRunnerCleansTempDirectoryOnCommandFailure(t *testing.T) {
	root := t.TempDir()
	runner := ProcessRunner{Limits: Limits{MaxSizeBytes: 200, MaxDuration: 10 * time.Second, MaxHeight: 720, Timeout: time.Second, TempRoot: root, YTDLP: "missing-yt-dlp", Proxy: "http://proxy:8080"}}
	if _, _, err := runner.Download(context.Background(), "https://www.youtube.com/watch?v=x"); err == nil {
		t.Fatal("expected unavailable binary error")
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("temporary directories were not cleaned: %v", entries)
	}
}

func TestProcessRunnerManagedDownloadOwnsFileUntilCleanup(t *testing.T) {
	root := t.TempDir()
	runner := ProcessRunner{Limits: Limits{MaxSizeBytes: 64, MaxDuration: time.Minute, MaxHeight: 720, TempRoot: root, Proxy: "http://proxy:8080"}}
	path, cleanup, err := runner.downloadWithCommand(context.Background(), func(_ context.Context, _ string, args ...string) ([]byte, error) {
		for i := 0; i < len(args)-1; i++ {
			if args[i] == "-o" {
				output := args[i+1]
				output = strings.Replace(output, "%(ext)s", "mp4", 1)
				return nil, os.WriteFile(output, []byte("video"), 0600)
			}
		}
		t.Fatal("output argument missing")
		return nil, nil
	}, "https://www.youtube.com/watch?v=x")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("managed file missing before cleanup: %v", err)
	}
	cleanup()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("managed file still exists after cleanup: %v", err)
	}
}

func TestProcessRunnerRejectsOversizedOutput(t *testing.T) {
	root := t.TempDir()
	runner := ProcessRunner{Limits: Limits{MaxSizeBytes: 4, MaxDuration: time.Minute, MaxHeight: 720, TempRoot: root, Proxy: "http://proxy:8080"}}
	_, _, err := runner.downloadWithCommand(context.Background(), func(_ context.Context, _ string, args ...string) ([]byte, error) {
		for i := 0; i < len(args)-1; i++ {
			if args[i] == "-o" {
				output := strings.Replace(args[i+1], "%(ext)s", "mp4", 1)
				return nil, os.WriteFile(output, []byte("oversized"), 0600)
			}
		}
		t.Fatal("output argument missing")
		return nil, nil
	}, "https://www.youtube.com/watch?v=x")
	if err == nil {
		t.Fatal("expected oversized output error")
	}
	entries, readErr := os.ReadDir(root)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("oversized output was not cleaned: %v", entries)
	}
}
