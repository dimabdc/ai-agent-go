package internal

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/adk/backend/local"
	"github.com/cloudwego/eino/adk/filesystem"
	"path/filepath"
)

type WrappedLocalBackend struct {
	backend filesystem.Backend
	baseDir string
}

func NewWrappedLocalBackend(ctx context.Context, baseDir string) (*WrappedLocalBackend, error) {
	// Создаём стандартный локальный бэкенд
	localBackend, err := local.NewBackend(ctx, &local.Config{})
	if err != nil {
		return nil, err
	}

	return &WrappedLocalBackend{
		backend: localBackend,
		baseDir: baseDir,
	}, nil
}

// LsInfo - список файлов и директорий
func (w *WrappedLocalBackend) LsInfo(
	ctx context.Context,
	req *filesystem.LsInfoRequest,
) ([]filesystem.FileInfo, error) {
	reqCopy := *req
	reqCopy.Path = filepath.Join(w.baseDir, req.Path)
	return w.backend.LsInfo(ctx, &reqCopy)
}

// Read - чтение файла
func (w *WrappedLocalBackend) Read(ctx context.Context, req *filesystem.ReadRequest) (*filesystem.FileContent, error) {
	reqCopy := *req
	reqCopy.FilePath = filepath.Join(w.baseDir, req.FilePath)
	return w.backend.Read(ctx, &reqCopy)
}

// GrepRaw - поиск по содержимому файлов
func (w *WrappedLocalBackend) GrepRaw(
	ctx context.Context,
	req *filesystem.GrepRequest,
) ([]filesystem.GrepMatch, error) {
	reqCopy := *req
	reqCopy.Path = filepath.Join(w.baseDir, req.Path)
	return w.backend.GrepRaw(ctx, &reqCopy)
}

// GlobInfo - поиск файлов по glob-паттерну
func (w *WrappedLocalBackend) GlobInfo(
	ctx context.Context,
	req *filesystem.GlobInfoRequest,
) ([]filesystem.FileInfo, error) {
	reqCopy := *req
	reqCopy.Path = filepath.Join(w.baseDir, req.Path)
	return w.backend.GlobInfo(ctx, &reqCopy)
}
func (w *WrappedLocalBackend) Write(ctx context.Context, req *filesystem.WriteRequest) error {
	return fmt.Errorf("write operation is disabled")
}

func (w *WrappedLocalBackend) Edit(ctx context.Context, req *filesystem.EditRequest) error {
	return fmt.Errorf("edit operation is disabled")
}
