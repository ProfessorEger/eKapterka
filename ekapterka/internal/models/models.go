// Package models определяет доменные сущности и shared-константы проекта.
// Эти структуры используются в bot/repository/seed слоях.
package models

// Файл содержит модели: Category, Item, Rental, UserState.

import "time"

// RootParentID — служебный идентификатор "виртуального" корня дерева категорий.
const RootParentID = "root"

// Константы ролей пользователя в системе.
const ADMIN = "admin"
const USER = "user"

// Category описывает узел дерева категорий снаряжения/инвентаря.
type Category struct {
	ID       string  `firestore:"id"`
	Title    string  `firestore:"title"`
	ParentID *string `firestore:"parent_id"`
	Order    int     `firestore:"order"`
	IsLeaf   bool    `firestore:"is_leaf"`
}

// Rental — период аренды конкретного предмета.
type Rental struct {
	ID          string    `firestore:"-"`
	ItemID      string    `firestore:"item_id"`
	Start       time.Time `firestore:"start"`
	End         time.Time `firestore:"end"`
	Description string    `firestore:"description"`
	UserID      int64     `firestore:"user_id"`
}

// Item — карточка предмета в каталоге.
// ID хранится как Firestore document ID и не сериализуется в поле документа.
type Item struct {
	ID          string    `firestore:"-"`
	Title       string    `firestore:"title"`
	Description string    `firestore:"description"`
	CategoryID  string    `firestore:"category_id"`
	Tags        []string  `firestore:"tags"`
	PhotoURLs   []string  `firestore:"photo_urls"`
	CreatedAt   time.Time `firestore:"created_at"`
	UpdatedAt   time.Time `firestore:"updated_at"`
}

// UserState хранит минимальное состояние пользователя для авторизации и UX.
type UserState struct {
	UserID    int64     `firestore:"id"`
	Role      string    `firestore:"role"`
	CreatedAt time.Time `firestore:"created_at"`
	MessageID int64     `firestore:"message_id"`
}
