package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	// DefaultMaxFileSize - максимальный размер файла в архиве (100MB)
	DefaultMaxFileSize = 100 * 1024 * 1024
	// DefaultMaxArchiveSize - максимальный размер всего архива (500MB)
	DefaultMaxArchiveSize = 500 * 1024 * 1024
)

// ArchiveExtractor - распаковщик tar.gz архивов
type ArchiveExtractor struct {
	// MaxFileSize - максимальный размер одного файла в архиве
	MaxFileSize int64
	// MaxArchiveSize - максимальный размер всего архива
	MaxArchiveSize int64
}

// NewArchiveExtractor создает новый экземпляр ArchiveExtractor с настройками по умолчанию
func NewArchiveExtractor() *ArchiveExtractor {
	return &ArchiveExtractor{
		MaxFileSize:    DefaultMaxFileSize,
		MaxArchiveSize: DefaultMaxArchiveSize,
	}
}

// Extract распаковывает tar.gz архив из reader в целевую директорию
func (e *ArchiveExtractor) Extract(r io.Reader, targetDir string) error {
	// Валидация входных данных
	if r == nil {
		return fmt.Errorf("reader cannot be nil")
	}

	if targetDir == "" {
		return fmt.Errorf("target directory cannot be empty")
	}

	// Создаем gzip reader
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	// Создаем tar reader
	tr := tar.NewReader(gzr)

	// Получаем абсолютный путь целевой директории
	absTargetDir, err := filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of target directory: %w", err)
	}

	// Счетчик общего размера архива
	var totalSize int64

	// Распаковка файлов
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar archive: %w", err)
		}

		// Проверка на path traversal
		cleanPath, err := ValidatePath(absTargetDir, header.Name)
		if err != nil {
			return fmt.Errorf("invalid path in archive: %w", err)
		}

		// Обработка директорий
		if header.Typeflag == tar.TypeDir {
			if err := e.createDirectory(cleanPath, header.Mode); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", cleanPath, err)
			}
			continue
		}

		// Обработка файлов
		if header.Typeflag == tar.TypeReg {
			// Проверка размера файла
			if header.Size > e.MaxFileSize {
				return fmt.Errorf("file %s exceeds maximum size limit (%d bytes)", header.Name, e.MaxFileSize)
			}

			// Проверка общего размера архива
			if totalSize+header.Size > e.MaxArchiveSize {
				return fmt.Errorf("archive exceeds maximum size limit (%d bytes)", e.MaxArchiveSize)
			}

			// Распаковка файла
			if err := e.extractFile(cleanPath, tr, header.Size); err != nil {
				return fmt.Errorf("failed to extract file %s: %w", header.Name, err)
			}

			totalSize += header.Size
		}

		// Игнорируем символические ссылки для безопасности
		if header.Typeflag == tar.TypeSymlink {
			return fmt.Errorf("symbolic links are not allowed in archive: %s", header.Name)
		}
	}

	return nil
}

// createDirectory создает директорию с указанными правами доступа
func (e *ArchiveExtractor) createDirectory(path string, mode int64) error {
	return os.MkdirAll(path, os.FileMode(mode))
}

// extractFile распаковывает файл из tar архива в целевую директорию
func (e *ArchiveExtractor) extractFile(path string, tr *tar.Reader, size int64) error {
	// Создаем родительские директории если нужно
	parentDir := filepath.Dir(path)
	if parentDir != "." && parentDir != "" {
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return err
		}
	}

	// Открываем файл для записи
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}
	defer file.Close()

	// Копируем данные из tar архива с ограничением размера
	lr := io.LimitReader(tr, size)
	if _, err := io.Copy(file, lr); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// ValidatePath проверяет безопасность пути и возвращает абсолютный путь
func ValidatePath(targetDir, archivePath string) (string, error) {
	// Нормализация пути
	cleanPath := filepath.Clean(archivePath)

	// Проверка на пустой путь
	if cleanPath == "" || cleanPath == "." {
		return targetDir, nil
	}

	// Проверка на ".." в начале пути (после нормализации это должно быть обработано, но для надежности)
	if strings.HasPrefix(cleanPath, "..") {
		return "", fmt.Errorf("path traversal attempt detected: %s", archivePath)
	}

	// Замена разделителей путей для кроссплатформенности
	cleanPath = filepath.ToSlash(cleanPath)

	// Проверка на ".." в середине или конце пути
	parts := strings.Split(cleanPath, "/")
	for _, part := range parts {
		if part == ".." {
			return "", fmt.Errorf("path traversal attempt detected: %s", archivePath)
		}
	}

	// Формирование полного пути
	finalPath := filepath.Join(targetDir, cleanPath)

	// Получаем абсолютный путь
	absFinalPath, err := filepath.Abs(finalPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Проверка что путь находится внутри целевой директории
	if !strings.HasPrefix(absFinalPath, targetDir+string(filepath.Separator)) && absFinalPath != targetDir {
		return "", fmt.Errorf("path traversal attempt detected: %s", archivePath)
	}

	return absFinalPath, nil
}
