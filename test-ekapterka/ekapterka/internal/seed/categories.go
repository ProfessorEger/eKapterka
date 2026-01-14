package seed

import (
	"ekapterka/internal/models"
)

func strPtr(s string) *string {
	return &s
}

var Categories = []models.Category{
	{
		ID:       "equipment",
		Title:    "Снаряжение",
		ParentID: nil,
		Path:     []string{"equipment"},
		Level:    0,
		Order:    10,
		IsLeaf:   false,
	},
	{
		ID:       "mountain",
		Title:    "Горное",
		ParentID: strPtr("equipment"),
		Path:     []string{"equipment", "mountain"},
		Level:    1,
		Order:    20,
		IsLeaf:   false,
	},
	{
		ID:       "ropes",
		Title:    "Веревки",
		ParentID: strPtr("mountain"),
		Path:     []string{"equipment", "mountain", "ropes"},
		Level:    2,
		Order:    30,
		IsLeaf:   true,
	},
}
