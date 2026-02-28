package repository

import (
	"context"
	"strconv"
	"time"

	"ekapterka/internal/models"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const usersCollection = "users"

func (c *Client) EnsureUserState(ctx context.Context, userID int64) error {
	docID := strconv.FormatInt(userID, 10)
	docRef := c.db.Collection(usersCollection).Doc(docID)

	_, err := docRef.Get(ctx)
	if err == nil {
		return nil
	}
	if status.Code(err) != codes.NotFound {
		return err
	}

	state := models.UserState{
		UserID:    userID,
		Role:      models.USER,
		CreatedAt: time.Now(),
		MessageID: 0,
	}

	_, err = docRef.Create(ctx, state)
	if status.Code(err) == codes.AlreadyExists {
		return nil
	}

	return err
}

func (c *Client) SetUserRole(ctx context.Context, userID int64, role string) error {
	docID := strconv.FormatInt(userID, 10)
	docRef := c.db.Collection(usersCollection).Doc(docID)

	_, err := docRef.Update(ctx, []firestore.Update{
		{Path: "role", Value: role},
	})
	if err == nil {
		return nil
	}
	if status.Code(err) != codes.NotFound {
		return err
	}

	if err := c.EnsureUserState(ctx, userID); err != nil {
		return err
	}

	_, err = docRef.Update(ctx, []firestore.Update{
		{Path: "role", Value: role},
	})
	return err
}
