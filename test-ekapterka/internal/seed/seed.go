package seed

import (
	"context"

	"cloud.google.com/go/firestore"
)

func SeedCategories(ctx context.Context, client *firestore.Client) {
	bw := client.BulkWriter(ctx)
	defer bw.End()

	for _, c := range Categories {
		ref := client.Collection("categories").Doc(c.ID)
		bw.Set(ref, c)
	}

	bw.Flush()
}
