package main

import (
	"context"
	"log"

	"test-ekapterka/internal/seed"

	"cloud.google.com/go/firestore"
)

func main() {
	ctx := context.Background()

	client, err := firestore.NewClient(ctx, "e-kapterka")
	if err != nil {
		log.Fatalf("firestore init failed: %v", err)
	}
	defer client.Close()

	seed.SeedCategories(ctx, client)

	log.Println("Categories seeded successfully")
}
