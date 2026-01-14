package repository

import (
	"context"
	"ekapterka/internal/models"

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

func (c *Client) GetChildCategories(ctx context.Context, parentID *string) ([]models.Category, error) {
	q := c.db.Collection("categories").Query

	if parentID == nil {
		q = q.Where("parent_id", "==", nil)
	} else {
		q = q.Where("parent_id", "==", *parentID)
	}

	iter := q.OrderBy("order", firestore.Asc).Documents(ctx)
	defer iter.Stop()

	var result []models.Category
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var c models.Category
		if err := doc.DataTo(&c); err != nil {
			return nil, err
		}
		result = append(result, c)
	}

	return result, nil
}
