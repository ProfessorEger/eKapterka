package seed

type Category struct {
	ID       string   `firestore:"id"`
	Title    string   `firestore:"title"`
	ParentID *string  `firestore:"parent_id"`
	Path     []string `firestore:"path"`
	Level    int      `firestore:"level"`
	Order    int      `firestore:"order"`
	IsLeaf   bool     `firestore:"is_leaf"`
}

// struct Items

// struct User
