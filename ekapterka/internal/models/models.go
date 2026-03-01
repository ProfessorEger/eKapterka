package models

import "time"

const RootParentID = "root"
const ADMIN = "admin"
const USER = "user"

type Category struct {
	ID       string   `firestore:"id"`
	Title    string   `firestore:"title"`
	ParentID *string  `firestore:"parent_id"`
	Path     []string `firestore:"path"`
	Level    int      `firestore:"level"`
	Order    int      `firestore:"order"`
	IsLeaf   bool     `firestore:"is_leaf"`
}

type Rental struct {
	Start       time.Time `firestore:"start"`
	End         time.Time `firestore:"end"`
	Description string    `firestore:"description"`
}

type Item struct {
	ID          string    `firestore:"-"`
	Title       string    `firestore:"title"`
	Description string    `firestore:"description"`
	CategoryID  string    `firestore:"category_id"`
	Tags        []string  `firestore:"tags"`
	PhotoURLs   []string  `firestore:"photo_urls"`
	CreatedAt   time.Time `firestore:"created_at"`
	UpdatedAt   time.Time `firestore:"updated_at"`
	Rentals     []Rental  `firestore:"rentals"`
}

type UserState struct {
	UserID    int64     `firestore:"id"`
	Role      string    `firestore:"role"`
	CreatedAt time.Time `firestore:"created_at"`
	MessageID int64     `firestore:"message_id"`
}
