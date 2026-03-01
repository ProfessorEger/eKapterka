package main

import (
	"context"
	"log"

	"ekapterka/internal/config"
	"ekapterka/internal/seed"

	"cloud.google.com/go/firestore"
)

func main() {
	ctx := context.Background()
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

	log.Println("Categories seeded successfully")
}
