package product

import "time"

type Product struct {
	ID                string    `db:"id" json:"id"`
	CategoryID        string    `db:"category_id" json:"category_id"`
	Name              string    `db:"name" json:"name"`
	Price             float64   `db:"price" json:"price"`
	Stock             int       `db:"stock" json:"stock"`
	LowStockThreshold int       `db:"low_stock_threshold" json:"low_stock_threshold"`
	ImageURL          *string   `db:"image_url" json:"image_url"`
	IsAvailable       bool      `db:"is_available" json:"is_available"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}

type CreateProductRequest struct {
	CategoryID        string  `json:"category_id" binding:"required"`
	Name              string  `json:"name" binding:"required"`
	Price             float64 `json:"price" binding:"required,min=0"`
	Stock             int     `json:"stock"`
	LowStockThreshold int     `json:"low_stock_threshold"`
	ImageURL          *string `json:"image_url"`
	IsAvailable       bool    `json:"is_available"`
}

type UpdateProductRequest struct {
	CategoryID        string  `json:"category_id" binding:"required"`
	Name              string  `json:"name" binding:"required"`
	Price             float64 `json:"price" binding:"required,min=0"`
	Stock             int     `json:"stock"`
	LowStockThreshold int     `json:"low_stock_threshold"`
	ImageURL          *string `json:"image_url"`
	IsAvailable       bool    `json:"is_available"`
}
