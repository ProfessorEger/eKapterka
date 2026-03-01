// Package seed содержит исходные данные и утилиты для первичного
// заполнения справочных коллекций в Firestore.
package seed

// Файл отвечает за применение seed-набора категорий в базу.

import (
	"context"

	"cloud.google.com/go/firestore"
)

// SeedCategories записывает предопределенный набор категорий в Firestore.
// Используется idempotent upsert через Set по фиксированным document ID.
func SeedCategories(ctx context.Context, client *firestore.Client) error {
	for _, c := range Categories {
		_, err := client.
			Collection("categories").
			Doc(c.ID).
			Set(ctx, c)
		if err != nil {
			return err
		}
	}
	return nil
}
