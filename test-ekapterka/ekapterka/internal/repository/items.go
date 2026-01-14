package repository

import (
	"context"
	"time"

	"ekapterka/internal/models"
)

/*
func (c *Client) AddItem(ctx context.Context, item models.Item) error {
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	_, err := c.db.Collection("items").Doc(item.ID).Set(ctx, item)
	return err
}
*/

func (c *Client) AddItem(ctx context.Context, item models.Item) error {
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	_, _, err := c.db.Collection("items").Add(ctx, item)
	return err
}
