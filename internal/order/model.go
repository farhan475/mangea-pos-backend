package order

import "time"

type OrderStatus string

const (
	StatusPending  OrderStatus = "pending"
	StatusCooking  OrderStatus = "cooking"
	StatusReady    OrderStatus = "ready"
	StatusPaid     OrderStatus = "paid"
	StatusCanceled OrderStatus = "cancelled"
)

type SyncStatus string

const (
	SyncStatusSynced  SyncStatus = "synced"
	SyncStatusPending SyncStatus = "pending"
)

type Order struct {
	ID            string      `db:"id" json:"id"`
	UserID        *string     `db:"user_id" json:"user_id"`
	CustomerName  *string     `db:"customer_name" json:"customer_name"`
	TableNumber   *string     `db:"table_number" json:"table_number"`
	TotalAmount   float64     `db:"total_amount" json:"total_amount"`
	Status        string      `db:"status" json:"status"`
	PaymentMethod *string     `db:"payment_method" json:"payment_method"`
	PaidAmount    *float64    `db:"paid_amount" json:"paid_amount"`
	ChangeAmount  *float64    `db:"change_amount" json:"change_amount"`
	SyncStatus    string      `db:"sync_status" json:"sync_status"`
	CreatedAt     time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time   `db:"updated_at" json:"updated_at"`
	Items         []OrderItem `db:"-" json:"items"`
}

type OrderItem struct {
	ID          string  `db:"id" json:"id"`
	OrderID     string  `db:"order_id" json:"order_id"`
	ProductID   string  `db:"product_id" json:"product_id"`
	ProductName string  `db:"product_name" json:"product_name"`
	Price       float64 `db:"price" json:"price"`
	Quantity    int     `db:"quantity" json:"quantity"`
	Subtotal    float64 `db:"subtotal" json:"subtotal"`
}

type CreateOrderRequest struct {
	ID           string                   `json:"id"` // optional: client-generated UUID to keep IDs aligned across devices
	UserID       *string                  `json:"user_id"`
	CustomerName *string                  `json:"customer_name"`
	TableNumber  *string                  `json:"table_number"`
	Items        []CreateOrderItemRequest `json:"items" binding:"required,min=1"`
}

type CreateOrderItemRequest struct {
	ID          string  `json:"id"`
	ProductID   string  `json:"product_id" binding:"required"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"` // ignored: price is resolved server-side from products table
	Quantity    int     `json:"quantity" binding:"required,min=1"`
	Subtotal    float64 `json:"subtotal"` // ignored: computed server-side
}

// PaymentRequest represents a payment for an existing order
type PaymentRequest struct {
	PaymentMethod string  `json:"payment_method" binding:"required,oneof=cash card ewallet qris"`
	PaidAmount    float64 `json:"paid_amount" binding:"required,min=0"`
}

type UpdateOrderRequest struct {
	UserID        *string  `json:"user_id"`
	CustomerName  *string  `json:"customer_name"`
	TableNumber   *string  `json:"table_number"`
	TotalAmount   float64  `json:"total_amount" binding:"required,min=0"`
	Status        string   `json:"status"`
	PaymentMethod *string  `json:"payment_method"`
	PaidAmount    *float64 `json:"paid_amount"`
	ChangeAmount  *float64 `json:"change_amount"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
