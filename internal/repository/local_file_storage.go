package repository

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type LocalFileStorage struct {
	baseDir string
}

func NewLocalFileStorage(baseDir string) (*LocalFileStorage, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("Создание файловой дерриктории: %w", err)
	}
	return &LocalFileStorage{baseDir: baseDir}, nil
}

func (s *LocalFileStorage) Save(ctx context.Context, filename string, content io.Reader) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	ext := filepath.Ext(filename)
	diskName := uuid.NewString() + ext
	fullPath := filepath.Join(s.baseDir, diskName)

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("Создание файла: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, content); err != nil {
		return "", fmt.Errorf("Запись файла: %w", err)
	}

	return fullPath, nil
}

func (s *LocalFileStorage) Delete(ctx context.Context, path string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		// Если файла нет на диске — считаем это успехом (идемпотентность)
		return fmt.Errorf("Удаление файла: %w", err)
	}
	return nil
}
