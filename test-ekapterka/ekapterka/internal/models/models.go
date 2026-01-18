package models

import "time"

type Category struct {
	ID       string   `firestore:"id"`
	Title    string   `firestore:"title"`
	ParentID *string  `firestore:"parent_id"`
	Path     []string `firestore:"path"`
	Level    int      `firestore:"level"`
	Order    int      `firestore:"order"`
	IsLeaf   bool     `firestore:"is_leaf"`
}

type Item struct {
	//ID           string    `firestore:"id"`
	Title        string    `firestore:"title"`
	Description  string    `firestore:"description"`
	CategoryID   string    `firestore:"category_id"`
	CategoryPath []string  `firestore:"category_path"`
	Tags         []string  `firestore:"tags"`
	CreatedAt    time.Time `firestore:"created_at"`
	UpdatedAt    time.Time `firestore:"updated_at"`
}

type UserState struct {
	UserID    int64
	Screen    string            // "categories", "products", "product"
	Params    map[string]string // cat_id, product_id, page
	History   []UserScreen      // стек
	UpdatedAt time.Time
}

type UserScreen struct {
	Screen string
	Params map[string]string
}
