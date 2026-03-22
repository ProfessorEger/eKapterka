package repository

// Файл содержит операции работы с предметами:
// создание, чтение, редактирование и удаление.

import (
	"context"
	"log"
	"time"

	"ekapterka/internal/models"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// itemFromDoc преобразует Firestore документ в доменную модель Item.
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

// AddItem создает новый предмет с автоматическими created_at/updated_at.
func (c *Client) AddItem(ctx context.Context, item models.Item) error {
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	_, _, err := c.db.Collection("items").Add(ctx, item)
	return err
}

// GetItemsByCategoryPage возвращает страницу предметов категории.
// Для определения hasNext используется выборка limit+1.
// Если запрос с сортировкой не проходит (например, из-за индекса),
// применяется fallback-запрос без OrderBy.
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
		OrderBy("title", firestore.Asc).
		Offset(offset).
		Limit(limit + 1)

	items, hasNext, err := c.readItemsQuery(ctx, q, limit)
	if err == nil {
		return items, hasNext, nil
	}

	log.Printf("items ordered query failed: %v", err)
	return nil, false, err
}

// readItemsQuery исполняет Firestore query и вычисляет флаг следующей страницы.
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


// GetItemByID возвращает предмет по document ID.
func (c *Client) GetItemByID(ctx context.Context, id string) (*models.Item, error) {
	doc, err := c.db.Collection("items").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}

	item := itemFromDoc(doc)

	return &item, nil
}

// GetItemsByIDs возвращает предметы по списку ID.
func (c *Client) GetItemsByIDs(ctx context.Context, ids []string) ([]models.Item, error) {
	if len(ids) == 0 {
		return []models.Item{}, nil
	}

	refs := make([]*firestore.DocumentRef, 0, len(ids))
	for _, id := range ids {
		refs = append(refs, c.db.Collection("items").Doc(id))
	}

	docs, err := c.db.GetAll(ctx, refs)
	if err != nil {
		return nil, err
	}

	items := make([]models.Item, 0, len(docs))
	for _, doc := range docs {
		if doc == nil || !doc.Exists() {
			continue
		}
		items = append(items, itemFromDoc(doc))
	}

	return items, nil
}

// UpdateItem обновляет базовые поля предмета и время обновления.
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

// DeleteItemByID удаляет документ предмета по ID.
func (c *Client) DeleteItemByID(ctx context.Context, id string) error {
	_, err := c.db.Collection("items").Doc(id).Delete(ctx)
	return err
}
