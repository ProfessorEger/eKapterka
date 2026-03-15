package repository

// Файл содержит операции работы с предметами:
// создание, чтение, редактирование, удаление и управление периодами аренды.

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"ekapterka/internal/models"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// itemFromDoc преобразует Firestore документ в доменную модель Item.
// Здесь также поддерживается обратная совместимость по полю rental_periods.
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
	parseRentals := func(rawRentals []interface{}) {
		item.Rentals = make([]models.Rental, 0, len(rawRentals))
		for _, rawRental := range rawRentals {
			rentalMap, ok := rawRental.(map[string]interface{})
			if !ok {
				continue
			}

			start, okStart := rentalMap["start"].(time.Time)
			end, okEnd := rentalMap["end"].(time.Time)
			if !okStart || !okEnd {
				continue
			}

			description := ""
			if v, ok := rentalMap["description"].(string); ok {
				description = v
			}

			var userID int64
			switch v := rentalMap["user_id"].(type) {
			case int64:
				userID = v
			case int:
				userID = int64(v)
			case float64:
				userID = int64(v)
			}

			username := ""
			if v, ok := rentalMap["username"].(string); ok {
				username = v
			}

			itemID := ""
			if v, ok := rentalMap["item_id"].(string); ok {
				itemID = v
			}

			item.Rentals = append(item.Rentals, models.Rental{
				ItemID:      itemID,
				Start:       start,
				End:         end,
				Description: description,
				UserID:      userID,
				Username:    username,
			})
		}
	}

	if rawRentals, ok := data["rentals"].([]interface{}); ok {
		parseRentals(rawRentals)
	} else if rawRentalPeriods, ok := data["rental_periods"].([]interface{}); ok {
		// Backward compatibility with old field name.
		parseRentals(rawRentalPeriods)
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

	rentals, err := c.GetRentalsByItemID(ctx, id)
	if err != nil {
		log.Printf("get rentals for item %s failed: %v", id, err)
		return &item, nil
	}
	if len(rentals) > 0 {
		item.Rentals = mergeRentals(item.Rentals, rentals)
	}

	return &item, nil
}

func mergeRentals(existing []models.Rental, incoming []models.Rental) []models.Rental {
	if len(existing) == 0 {
		return incoming
	}
	if len(incoming) == 0 {
		return existing
	}

	seen := make(map[string]struct{}, len(existing)+len(incoming))
	result := make([]models.Rental, 0, len(existing)+len(incoming))

	addRental := func(r models.Rental) {
		key := rentalKey(r)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		result = append(result, r)
	}

	for _, r := range existing {
		addRental(r)
	}
	for _, r := range incoming {
		addRental(r)
	}

	return result
}

func rentalKey(r models.Rental) string {
	return r.Start.UTC().Format(time.RFC3339Nano) + "|" +
		r.End.UTC().Format(time.RFC3339Nano) + "|" +
		strconv.FormatInt(r.UserID, 10) + "|" +
		strings.ToLower(strings.TrimSpace(r.Username)) + "|" +
		strings.TrimSpace(r.Description)
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

// UpdateItemRentals полностью перезаписывает массив rentals.
// Используется, например, при удалении аренды по номеру.
func (c *Client) UpdateItemRentals(ctx context.Context, id string, rentals []models.Rental) error {
	_, err := c.db.Collection("items").Doc(id).Update(ctx, []firestore.Update{
		{Path: "rentals", Value: rentals},
		{Path: "updated_at", Value: time.Now()},
	})
	return err
}
