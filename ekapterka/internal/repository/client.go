package repository

import (
	"context"
	"log"
	"strings"

	"cloud.google.com/go/firestore"
)

type Client struct {
	db *firestore.Client
}

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

func (c *Client) Close() {
	c.db.Close()
}
