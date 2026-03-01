// Package storage содержит адаптеры файлового хранилища.
// Текущая реализация работает с Google Cloud Storage.
package storage

// Файл реализует GCS-клиент: загрузку/удаление объектов и формирование public URL.

import (
	"context"
	"fmt"
	"io"
	"log"
	"path"
	"strings"

	"cloud.google.com/go/storage"
)

// Storage — минимальный контракт для работы с файловым хранилищем.
// Нужен, чтобы изолировать bot-логику от конкретной реализации GCS.
type Storage interface {
	Upload(ctx context.Context, objectPath string, r io.Reader, contentType string) error
	Delete(ctx context.Context, objectPath string) error
	PublicURL(objectPath string) string
}

// GCS — реализация Storage поверх Google Cloud Storage.
type GCS struct {
	client     *storage.Client
	bucketName string
}

// NewGCS создает клиент GCS для заданного bucket.
func NewGCS(ctx context.Context, bucketName string) *GCS {
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("storage init failed: %v", err)
	}

	return &GCS{
		client:     client,
		bucketName: bucketName,
	}
}

// Upload загружает объект в bucket по указанному пути.
func (g *GCS) Upload(ctx context.Context, objectPath string, r io.Reader, contentType string) error {
	objectPath = cleanPath(objectPath)

	w := g.client.
		Bucket(g.bucketName).
		Object(objectPath).
		NewWriter(ctx)

	w.ContentType = contentType

	if _, err := io.Copy(w, r); err != nil {
		_ = w.Close()
		return fmt.Errorf("copy to gcs: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("close writer: %w", err)
	}

	return nil
}

// Delete удаляет объект из bucket.
func (g *GCS) Delete(ctx context.Context, objectPath string) error {
	objectPath = cleanPath(objectPath)

	err := g.client.
		Bucket(g.bucketName).
		Object(objectPath).
		Delete(ctx)

	if err != nil {
		return fmt.Errorf("delete object: %w", err)
	}

	return nil
}

// PublicURL возвращает публичный URL объекта в стандартном формате GCS.
func (g *GCS) PublicURL(objectPath string) string {
	objectPath = cleanPath(objectPath)
	return fmt.Sprintf(
		"https://storage.googleapis.com/%s/%s",
		g.bucketName,
		objectPath,
	)
}

// cleanPath нормализует путь объекта (убирает лишние префиксы/пробелы).
func cleanPath(p string) string {
	p = strings.TrimSpace(p)
	p = strings.TrimPrefix(p, "/")
	return path.Clean(p)
}
