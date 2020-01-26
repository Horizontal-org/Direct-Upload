package application

import (
	"context"
	"errors"
	"io"
)

var (
	ErrNoUserCtx = errors.New("no user in context")
	ErrConflict  = errors.New("conflict")
)

type FileInfo struct {
	Size int64
}

type FileStore interface {
	GetFileInfo(ctx context.Context, fileName string) (*FileInfo, error)
	AppendFile(ctx context.Context, file string, data io.ReadCloser) error
	CloseFile(ctx context.Context, file string) error
}
