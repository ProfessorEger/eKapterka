package repository

// Файл содержит операции работы с арендой в отдельной коллекции rentals.

import (
	"context"
	"log"
	"time"

	"ekapterka/internal/models"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

const rentalsCollection = "rentals"

// rentalFromDoc преобразует документ аренды в доменную модель Rental.
func rentalFromDoc(doc *firestore.DocumentSnapshot) models.Rental {
	data := doc.Data()
	rental := models.Rental{
		ID: doc.Ref.ID,
	}

	if v, ok := data["item_id"].(string); ok {
		rental.ItemID = v
	}
	if v, ok := data["start"].(time.Time); ok {
		rental.Start = v
	}
	if v, ok := data["end"].(time.Time); ok {
		rental.End = v
	}
	if v, ok := data["description"].(string); ok {
		rental.Description = v
	}
	if v, ok := data["username"].(string); ok {
		rental.Username = v
	}
	switch v := data["user_id"].(type) {
	case int64:
		rental.UserID = v
	case int:
		rental.UserID = int64(v)
	case float64:
		rental.UserID = int64(v)
	}

	return rental
}

// AddRental создает документ аренды и обновляет updated_at предмета.
func (c *Client) AddRental(ctx context.Context, itemID string, rental models.Rental) error {
	rental.ItemID = itemID
	if _, _, err := c.db.Collection(rentalsCollection).Add(ctx, rental); err != nil {
		return err
	}

	_, err := c.db.Collection("items").Doc(itemID).Update(ctx, []firestore.Update{
		{Path: "updated_at", Value: time.Now()},
	})
	return err
}

// GetRentalsByItemID возвращает аренды по ID предмета.
func (c *Client) GetRentalsByItemID(ctx context.Context, itemID string) ([]models.Rental, error) {
	iter := c.db.Collection(rentalsCollection).Where("item_id", "==", itemID).Documents(ctx)
	defer iter.Stop()

	rentals := make([]models.Rental, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("get rentals by item error: %v", err)
			return nil, err
		}

		rentals = append(rentals, rentalFromDoc(doc))
	}

	return rentals, nil
}

// GetRentalsByUserID возвращает аренды по Telegram user_id.
func (c *Client) GetRentalsByUserID(ctx context.Context, userID int64) ([]models.Rental, error) {
	iter := c.db.Collection(rentalsCollection).Where("user_id", "==", userID).Documents(ctx)
	defer iter.Stop()

	rentals := make([]models.Rental, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("get rentals by user error: %v", err)
			return nil, err
		}

		rentals = append(rentals, rentalFromDoc(doc))
	}

	return rentals, nil
}

// GetRentalByID возвращает аренду по document ID.
func (c *Client) GetRentalByID(ctx context.Context, id string) (models.Rental, error) {
	doc, err := c.db.Collection(rentalsCollection).Doc(id).Get(ctx)
	if err != nil {
		return models.Rental{}, err
	}

	return rentalFromDoc(doc), nil
}

// DeleteRental удаляет аренду и обновляет updated_at предмета.
func (c *Client) DeleteRental(ctx context.Context, rental models.Rental) error {
	if rental.ID == "" {
		return nil
	}

	if _, err := c.db.Collection(rentalsCollection).Doc(rental.ID).Delete(ctx); err != nil {
		return err
	}

	if rental.ItemID == "" {
		return nil
	}

	_, err := c.db.Collection("items").Doc(rental.ItemID).Update(ctx, []firestore.Update{
		{Path: "updated_at", Value: time.Now()},
	})
	return err
}
