// Package repository инкапсулирует доступ к Firestore
// и предоставляет прикладные CRUD-операции для доменных сущностей.
package repository

// Файл содержит базовую инициализацию и жизненный цикл Firestore-клиента.

import (
	"context"
	"log"
	"strings"

	"cloud.google.com/go/firestore"
)

// Client инкапсулирует доступ к Firestore.
type Client struct {
	db *firestore.Client
}

// NewClient создает Firestore-клиент для конкретного GCP проекта.
func NewClient(ctx context.Context, projectID string) *Client {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		log.Fatal("firestore init failed: project ID is empty")
	}

	db, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("firestore init failed: %v", err)
	}

	return &Client{db: db}
}

// Close закрывает сетевые ресурсы клиента Firestore.
func (c *Client) Close() {
	c.db.Close()
}
