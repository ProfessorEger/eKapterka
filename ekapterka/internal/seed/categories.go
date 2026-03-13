package seed

// Файл содержит фиксированный каталог категорий для сидирования.
// Это "источник истины" для структуры дерева категорий в Firestore.

import (
	"ekapterka/internal/models"
)

// strPtr удобен для краткой инициализации parent_id в seed-структурах.
func strPtr(s string) *string {
	return &s
}

// Categories — эталонное дерево категорий каталога.
// Важно:
// - ID должен быть стабильным (используется в командах /add, /edit и callback payload).
// - Для корневых узлов ParentID = models.RootParentID.

var Categories = []models.Category{
	{
		ID:       "tents",
		Title:    "Палатки",
		ParentID: strPtr(models.RootParentID),
		Order:    10,
		IsLeaf:   true,
	},
	{
		ID:       "backpacks_normal",
		Title:    "Рюкзаки",
		ParentID: strPtr(models.RootParentID),
		Order:    20,
		IsLeaf:   true,
	},
	{
		ID:       "sleeping_bags",
		Title:    "Спальные мешки",
		ParentID: strPtr(models.RootParentID),
		Order:    30,
		IsLeaf:   true,
	},
	{
		ID:       "axes_saws",
		Title:    "Топоры/пилы",
		ParentID: strPtr(models.RootParentID),
		Order:    40,
		IsLeaf:   true,
	},
}
