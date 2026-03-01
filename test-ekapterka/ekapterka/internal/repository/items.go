package repository

import (
	"context"
	"log"
	"time"

	"ekapterka/internal/models"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

func itemFromDoc(doc *firestore.DocumentSnapshot) models.Item {
	data := doc.Data()
	item := models.Item{
		ID: doc.Ref.ID,
	}

	if v, ok := data["title"].(string); ok {
		item.Title = v
	}
	if v, ok := data["description"].(string); ok {
		item.Description = v
	}
	if v, ok := data["category_id"].(string); ok {
		item.CategoryID = v
	}
	if v, ok := data["created_at"].(time.Time); ok {
		item.CreatedAt = v
	}
	if v, ok := data["updated_at"].(time.Time); ok {
		item.UpdatedAt = v
	}

	if rawPath, ok := data["category_path"].([]interface{}); ok {
		item.CategoryPath = make([]string, 0, len(rawPath))
		for _, p := range rawPath {
			if s, ok := p.(string); ok {
				item.CategoryPath = append(item.CategoryPath, s)
			}
		}
	}

	if rawTags, ok := data["tags"].([]interface{}); ok {
		item.Tags = make([]string, 0, len(rawTags))
		for _, t := range rawTags {
			if s, ok := t.(string); ok {
				item.Tags = append(item.Tags, s)
			}
		}
	}
	if rawPhotoURLs, ok := data["photo_urls"].([]interface{}); ok {
		item.PhotoURLs = make([]string, 0, len(rawPhotoURLs))
		for _, p := range rawPhotoURLs {
			if s, ok := p.(string); ok {
				item.PhotoURLs = append(item.PhotoURLs, s)
			}
		}
	}

	return item
}

func (c *Client) AddItem(ctx context.Context, item models.Item) error {
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	_, _, err := c.db.Collection("items").Add(ctx, item)
	return err
}

func (c *Client) GetItemsByCategoryPage(ctx context.Context, categoryID string, page, limit int) ([]models.Item, bool, error) {
	if page < 0 {
		page = 0
	}
	if limit <= 0 {
		limit = 10
	}

	offset := page * limit

	q := c.db.Collection("items").Query.
		Where("category_id", "==", categoryID).
		OrderBy("created_at", firestore.Desc).
		Offset(offset).
		Limit(limit + 1)

	items, hasNext, err := c.readItemsQuery(ctx, q, limit)
	if err == nil {
		return items, hasNext, nil
	}

	log.Printf("items ordered query failed, fallback without order: %v", err)

	fallbackQ := c.db.Collection("items").Query.
		Where("category_id", "==", categoryID).
		Offset(offset).
		Limit(limit + 1)

	return c.readItemsQuery(ctx, fallbackQ, limit)
}

func (c *Client) readItemsQuery(ctx context.Context, q firestore.Query, limit int) ([]models.Item, bool, error) {
	iter := q.Documents(ctx)
	defer iter.Stop()

	items := make([]models.Item, 0, limit+1)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("get items by category error: %v", err)
			return nil, false, err
		}

		items = append(items, itemFromDoc(doc))
	}

	hasNext := len(items) > limit
	if hasNext {
		items = items[:limit]
	}

	return items, hasNext, nil
}

func (c *Client) GetItemByID(ctx context.Context, id string) (*models.Item, error) {
	doc, err := c.db.Collection("items").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}

	item := itemFromDoc(doc)

	return &item, nil
}

func (c *Client) UpdateItem(
	ctx context.Context,
	id string,
	title string,
	categoryID string,
	description string,
	photoURLs []string,
) error {
	_, err := c.db.Collection("items").Doc(id).Update(ctx, []firestore.Update{
		{Path: "title", Value: title},
		{Path: "category_id", Value: categoryID},
		{Path: "description", Value: description},
		{Path: "photo_urls", Value: photoURLs},
		{Path: "updated_at", Value: time.Now()},
	})
	return err
}

func (c *Client) DeleteItemByID(ctx context.Context, id string) error {
	_, err := c.db.Collection("items").Doc(id).Delete(ctx)
	return err
}
