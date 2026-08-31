package dashboard

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

type Metrics struct {
	NewOrdersCount     int     `json:"new_orders_count"`
	TotalOrdersToday   int     `json:"total_orders_today"`
	OrdersGrowthPercent *float64 `json:"orders_growth_percent"`
	WaitingListCount   int     `json:"waiting_list_count"`
}

func (r *Repository) GetMetrics() (*Metrics, error) {
	metrics := &Metrics{}

	// New orders (pending in last 1 hour)
	err := r.db.Get(&metrics.NewOrdersCount, `
		SELECT COUNT(*) FROM orders
		WHERE status = 'pending' AND created_at >= DATE_SUB(NOW(), INTERVAL 1 HOUR)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get new orders count: %w", err)
	}

	// Total orders today (not cancelled)
	err = r.db.Get(&metrics.TotalOrdersToday, `
		SELECT COUNT(*) FROM orders
		WHERE DATE(created_at) = CURDATE() AND status != 'cancelled'
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get total orders today: %w", err)
	}

	// Growth percent (today vs yesterday)
	var yesterdayCount int
	err = r.db.Get(&yesterdayCount, `
		SELECT COUNT(*) FROM orders
		WHERE DATE(created_at) = DATE_SUB(CURDATE(), INTERVAL 1 DAY) AND status != 'cancelled'
	`)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get yesterday orders count: %w", err)
	}

	if yesterdayCount > 0 {
		growth := float64((metrics.TotalOrdersToday - yesterdayCount)) / float64(yesterdayCount) * 100
		metrics.OrdersGrowthPercent = &growth
	}

	// Waiting list (occupied tables with orders not ready/paid/cancelled)
	query := "SELECT COUNT(*) FROM `tables` t JOIN orders o ON o.id = t.current_order_id WHERE t.status = 'occupied' AND o.status NOT IN ('ready', 'paid', 'cancelled')"
	err = r.db.Get(&metrics.WaitingListCount, query)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get waiting list count: %w", err)
	}

	return metrics, nil
}

type PopularDish struct {
	ID        string  `db:"id"         json:"id"`
	Name      string  `db:"name"       json:"name"`
	SoldCount int     `db:"sold_count" json:"sold_count"`
	ImageURL  *string `db:"image_url"  json:"image_url"`
}

func (r *Repository) GetPopularDishes(limit int) ([]PopularDish, error) {
	if limit <= 0 {
		limit = 5
	}

	var dishes []PopularDish
	query := `
		SELECT p.id, p.name, SUM(oi.quantity) as sold_count, p.image_url
		FROM order_items oi
		JOIN products p ON oi.product_id = p.id
		JOIN orders o ON oi.order_id = o.id
		WHERE DATE(o.created_at) = CURDATE()
		GROUP BY p.id, p.name, p.image_url
		ORDER BY sold_count DESC
		LIMIT ?
	`

	err := r.db.Select(&dishes, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular dishes: %w", err)
	}

	return dishes, nil
}

type StockAlert struct {
	ID               string `db:"id"                json:"id"`
	ProductName      string `db:"product_name"      json:"product_name"`
	AvailabilityNote string `db:"availability_note" json:"availability_note"`
}

func (r *Repository) GetOutOfStock() ([]StockAlert, error) {
	var alerts []StockAlert
	query := `
		SELECT id, name as product_name,
			CASE
				WHEN stock = 0 THEN 'Out of stock'
				WHEN is_available = FALSE THEN 'Not available'
				ELSE 'Low stock'
			END as availability_note
		FROM products
		WHERE is_available = FALSE OR stock = 0
		ORDER BY name ASC
	`

	err := r.db.Select(&alerts, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get out of stock items: %w", err)
	}

	return alerts, nil
}
