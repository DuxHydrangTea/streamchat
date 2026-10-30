package storage

import (
	"context"
	"io"
)

type Object struct {
	Key         string
	Size        int64
	ContentType string
}
type Store interface {
	Put(context.Context, string, io.Reader, string, int64) error
	Open(context.Context, string) (io.ReadCloser, error)
	Delete(context.Context, string) error
}
