// Package main содержит entrypoint для одноразового сидирования данных.
// Этот файл предназначен для заполнения Firestore коллекции categories
// исходным деревом категорий из internal/seed.
package main

import (
	"context"
	"log"

	"ekapterka/internal/config"
	"ekapterka/internal/seed"

	"cloud.google.com/go/firestore"
)

// main выполняет одноразовое сидирование коллекции categories.
// Используется для первичного заполнения дерева категорий в Firestore.
func main() {
	ctx := context.Background()

	// Проект Firestore задается через env, чтобы не хардкодить окружение.
	projectID := config.MustEnv("FIRESTORE_PROJECT_ID")

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("firestore init failed: %v", err)
	}
	defer client.Close()

	err = seed.SeedCategories(ctx, client)
	if err != nil {
		log.Fatalf("failed to seed categories: %v", err)
	}

	// Важно: Set в seed.SeedCategories перезаписывает документы с теми же ID.
	log.Println("Categories seeded successfully")
}
