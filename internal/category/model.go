package category

import "time"

type Category struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type CreateCategoryRequest struct {
	Name string `json:"name" binding:"required"`
}
