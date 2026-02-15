package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"path"
	"strings"

	"cloud.google.com/go/storage"
)

type Storage interface {
	Upload(ctx context.Context, objectPath string, r io.Reader, contentType string) error
	Delete(ctx context.Context, objectPath string) error
	PublicURL(objectPath string) string
}

type GCS struct {
	client     *storage.Client
	bucketName string
}

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

func (g *GCS) PublicURL(objectPath string) string {
	objectPath = cleanPath(objectPath)
	return fmt.Sprintf(
		"https://storage.googleapis.com/%s/%s",
		g.bucketName,
		objectPath,
	)
}

func cleanPath(p string) string {
	p = strings.TrimSpace(p)
	p = strings.TrimPrefix(p, "/")
	return path.Clean(p)
}
