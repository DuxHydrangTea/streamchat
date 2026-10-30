package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Local struct{ Root string }

func (l Local) path(key string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(key))
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", errors.New("invalid object key")
	}
	root, err := filepath.Abs(l.Root)
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(filepath.Join(root, clean))
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(path, root+string(os.PathSeparator)) {
		return "", errors.New("invalid object key")
	}
	return path, nil
}
func (l Local) Put(ctx context.Context, key string, src io.Reader, _ string, _ int64) error {
	path, err := l.path(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := ctx.Err(); err != nil {
		return err
	}
	_, err = io.Copy(f, src)
	return err
}
func (l Local) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path, err := l.path(key)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}
func (l Local) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path, err := l.path(key)
	if err != nil {
		return err
	}
	return os.Remove(path)
}
