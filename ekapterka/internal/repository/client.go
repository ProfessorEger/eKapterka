package repository

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
)

type Client struct {
	db *firestore.Client
}

func NewClient(ctx context.Context) *Client {
	db, err := firestore.NewClient(ctx, "e-kapterka")
	if err != nil {
		log.Fatalf("firestore init failed: %v", err)
	}

	return &Client{db: db}
}

func (c *Client) Close() {
	c.db.Close()
}
