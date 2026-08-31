package table

import "time"

type TableStatus string

const (
	StatusAvailable TableStatus = "available"
	StatusOccupied  TableStatus = "occupied"
	StatusReserved  TableStatus = "reserved"
)

type Table struct {
	ID             string     `db:"id" json:"id"`
	TableNumber    string     `db:"table_number" json:"table_number"`
	Capacity       int        `db:"capacity" json:"capacity"`
	Status         string     `db:"status" json:"status"`
	CurrentOrderID *string    `db:"current_order_id" json:"current_order_id"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
}

type CreateTableRequest struct {
	TableNumber string `json:"table_number" binding:"required"`
	Capacity    int    `json:"capacity" binding:"required,min=1"`
	Status      string `json:"status"`
}

type UpdateTableRequest struct {
	TableNumber string `json:"table_number" binding:"required"`
	Capacity    int    `json:"capacity" binding:"required,min=1"`
}

type UpdateTableStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
