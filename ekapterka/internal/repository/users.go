package repository

// Файл содержит операции управления состоянием пользователя:
// автосоздание user-state, чтение роли и изменение роли.

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

// EnsureUserState гарантирует существование user-документа.
// Если пользователь встречается впервые, создается запись с ролью USER.
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

// SetUserRole устанавливает роль пользователя.
// Если документа нет, сначала создается базовое состояние.
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

// GetUserRole возвращает роль пользователя.
// При отсутствии документа создается состояние по умолчанию (role=user).
func (c *Client) GetUserRole(ctx context.Context, userID int64) (string, error) {
	docID := strconv.FormatInt(userID, 10)
	docRef := c.db.Collection(usersCollection).Doc(docID)

	doc, err := docRef.Get(ctx)
	if err == nil {
		var state models.UserState
		if err := doc.DataTo(&state); err != nil {
			return "", err
		}
		return state.Role, nil
	}
	if status.Code(err) != codes.NotFound {
		return "", err
	}

	if err := c.EnsureUserState(ctx, userID); err != nil {
		return "", err
	}

	doc, err = docRef.Get(ctx)
	if err != nil {
		return "", err
	}

	var state models.UserState
	if err := doc.DataTo(&state); err != nil {
		return "", err
	}

	return state.Role, nil
}
