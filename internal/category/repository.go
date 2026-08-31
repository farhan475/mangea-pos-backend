package category

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"mangea-backend/internal/util"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List() ([]Category, error) {
	var categories []Category
	err := r.db.Select(&categories, "SELECT id, name, created_at FROM categories ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	return categories, nil
}

// Create adds a new category
func (r *Repository) Create(req CreateCategoryRequest) (*Category, error) {
	now := time.Now()
	category := Category{
		ID:        util.GenerateID(),
		Name:      req.Name,
		CreatedAt: now,
	}

	query := "INSERT INTO categories (id, name, created_at) VALUES (?, ?, ?)"
	if _, err := r.db.Exec(query, category.ID, category.Name, category.CreatedAt); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return &category, nil
}
