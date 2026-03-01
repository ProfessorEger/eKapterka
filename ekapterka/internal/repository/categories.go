package repository

import (
	"context"
	"ekapterka/internal/models"
	"log"
	"sort"
	"strings"

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
		q = q.Where("parent_id", "==", models.RootParentID)
	} else {
		q = q.Where("parent_id", "==", *parentID)
	}

	q = q.OrderBy("order", firestore.Asc)

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

func (c *Client) GetLeafCategories(ctx context.Context) ([]models.Category, error) {
	q := c.db.Collection("categories").Query.
		Where("is_leaf", "==", true)

	iter := q.Documents(ctx)
	defer iter.Stop()

	var result []models.Category
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("get leaf categories error: %v", err)
			return nil, err
		}

		var cat models.Category
		if err := doc.DataTo(&cat); err != nil {
			log.Printf("decode leaf category error: %v", err)
			return nil, err
		}

		result = append(result, cat)
	}

	sort.Slice(result, func(i, j int) bool {
		left := strings.Join(result[i].Path, "/")
		right := strings.Join(result[j].Path, "/")
		return left < right
	})

	return result, nil
}
