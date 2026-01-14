package repository

import (
	"context"
	"ekapterka/internal/models"
	"log"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

func (c *Client) GetCategoryByID(ctx context.Context, id string) (*models.Category, error) {
	doc, err := c.db.Collection("categories").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}

	var cat models.Category
	if err := doc.DataTo(&cat); err != nil {
		return nil, err
	}

	return &cat, nil
}

func (c *Client) GetChildCategories(ctx context.Context, parentID string) ([]models.Category, error) {
	q := c.db.Collection("categories").
		Where("parent_id", "==", parentID).
		OrderBy("order", firestore.Asc)

	iter := q.Documents(ctx)
	defer iter.Stop()

	var result []models.Category
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("get categories error: %v", err)
			return nil, err
		}

		var cat models.Category
		if err := doc.DataTo(&cat); err != nil {
			log.Printf("get categories error: %v", err)
			return nil, err
		}
		result = append(result, cat)
	}

	return result, nil
}
