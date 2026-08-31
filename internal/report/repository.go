package report

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

// taxRate mirrors the POS convention: stored total_amount includes tax,
// i.e. total = subtotal * (1 + taxRate). Tax is extracted from the inclusive total.
const taxRate = 0.10

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// DailyReport represents a daily sales report
type DailyReport struct {
	Date             string           `json:"date"`
	TotalOrders      int              `json:"total_orders"`
	CompletedOrders  int              `json:"completed_orders"`
	CancelledOrders  int              `json:"cancelled_orders"`
	TotalRevenue     float64          `json:"total_revenue"`
	TotalTax         float64          `json:"total_tax"`
	NetRevenue       float64          `json:"net_revenue"`
	TopProducts      []ProductSales   `json:"top_products"`
	HourlySales      []HourlySales    `json:"hourly_sales"`
	PaymentBreakdown []PaymentSummary `json:"payment_breakdown"`
}

// ProductSales represents sales data for a product
type ProductSales struct {
	ProductID    string  `db:"product_id"   json:"product_id"`
	ProductName  string  `db:"product_name" json:"product_name"`
	QuantitySold int     `db:"quantity_sold" json:"quantity_sold"`
	Revenue      float64 `db:"revenue"      json:"revenue"`
}

// HourlySales represents sales aggregated by hour
type HourlySales struct {
	Hour    int     `db:"hour"    json:"hour"`
	Orders  int     `db:"orders"  json:"orders"`
	Revenue float64 `db:"revenue" json:"revenue"`
}

// PaymentSummary represents payment method breakdown
type PaymentSummary struct {
	PaymentMethod string  `db:"payment_method" json:"payment_method"`
	Count         int     `db:"count"          json:"count"`
	Total         float64 `db:"total"          json:"total"`
}

// WeeklyReport represents a weekly summary
type WeeklyReport struct {
	WeekStart    string         `json:"week_start"`
	WeekEnd      string         `json:"week_end"`
	TotalOrders  int            `json:"total_orders"`
	TotalRevenue float64        `json:"total_revenue"`
	DailySummary []DailySummary `json:"daily_summary"`
}

// DailySummary is a simplified daily summary used in weekly reports
type DailySummary struct {
	Date    string  `db:"date"    json:"date"`
	Orders  int     `db:"orders"  json:"orders"`
	Revenue float64 `db:"revenue" json:"revenue"`
}

// GetDailyReport returns full daily report for a given date
func (r *Repository) GetDailyReport(date string) (*DailyReport, error) {
	report := &DailyReport{Date: date}

	// Total orders
	err := r.db.Get(&report.TotalOrders, `
		SELECT COUNT(*) FROM orders WHERE DATE(created_at) = ?`, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get total orders: %w", err)
	}

	// Completed orders
	err = r.db.Get(&report.CompletedOrders, `
		SELECT COUNT(*) FROM orders WHERE DATE(created_at) = ? AND status = 'paid'`, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get completed orders: %w", err)
	}

	// Cancelled orders
	err = r.db.Get(&report.CancelledOrders, `
		SELECT COUNT(*) FROM orders WHERE DATE(created_at) = ? AND status = 'cancelled'`, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get cancelled orders: %w", err)
	}

	// Revenue (from paid orders)
	err = r.db.Get(&report.TotalRevenue, `
		SELECT COALESCE(SUM(total_amount), 0) FROM orders 
		WHERE DATE(created_at) = ? AND status = 'paid'`, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get revenue: %w", err)
	}

	// Tax extracted from tax-inclusive totals (total = subtotal * 1.1)
	report.NetRevenue = report.TotalRevenue / (1 + taxRate)
	report.TotalTax = report.TotalRevenue - report.NetRevenue

	// Top products
	err = r.db.Select(&report.TopProducts, `
		SELECT 
			oi.product_id,
			oi.product_name,
			SUM(oi.quantity) as quantity_sold,
			SUM(oi.subtotal) as revenue
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.id
		WHERE DATE(o.created_at) = ? AND o.status = 'paid'
		GROUP BY oi.product_id, oi.product_name
		ORDER BY quantity_sold DESC
		LIMIT 10`, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get top products: %w", err)
	}
	if report.TopProducts == nil {
		report.TopProducts = []ProductSales{}
	}

	// Hourly sales
	err = r.db.Select(&report.HourlySales, `
		SELECT 
			HOUR(created_at) as hour,
			COUNT(*) as orders,
			COALESCE(SUM(total_amount), 0) as revenue
		FROM orders
		WHERE DATE(created_at) = ? AND status = 'paid'
		GROUP BY HOUR(created_at)
		ORDER BY hour ASC`, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get hourly sales: %w", err)
	}
	if report.HourlySales == nil {
		report.HourlySales = []HourlySales{}
	}

	// Payment method breakdown
	err = r.db.Select(&report.PaymentBreakdown, `
		SELECT 
			COALESCE(payment_method, 'unknown') as payment_method,
			COUNT(*) as count,
			COALESCE(SUM(total_amount), 0) as total
		FROM orders
		WHERE DATE(created_at) = ? AND status = 'paid'
		GROUP BY payment_method`, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment breakdown: %w", err)
	}
	if report.PaymentBreakdown == nil {
		report.PaymentBreakdown = []PaymentSummary{}
	}

	return report, nil
}

// GetWeeklyReport returns weekly summary for the week containing the given date
func (r *Repository) GetWeeklyReport(date string) (*WeeklyReport, error) {
	// Calculate week start (Monday) and end (Sunday)
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Get Monday of the week
	weekday := int(parsedDate.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekStart := parsedDate.AddDate(0, 0, -(weekday - 1))
	weekEnd := weekStart.AddDate(0, 0, 6)

	report := &WeeklyReport{
		WeekStart: weekStart.Format("2006-01-02"),
		WeekEnd:   weekEnd.Format("2006-01-02"),
	}

	// Total orders for the week
	err = r.db.Get(&report.TotalOrders, `
		SELECT COUNT(*) FROM orders 
		WHERE DATE(created_at) BETWEEN ? AND ? AND status != 'cancelled'`,
		report.WeekStart, report.WeekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get weekly orders: %w", err)
	}

	// Total revenue for the week
	err = r.db.Get(&report.TotalRevenue, `
		SELECT COALESCE(SUM(total_amount), 0) FROM orders 
		WHERE DATE(created_at) BETWEEN ? AND ? AND status = 'paid'`,
		report.WeekStart, report.WeekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get weekly revenue: %w", err)
	}

	// Daily breakdown
	err = r.db.Select(&report.DailySummary, `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as orders,
			COALESCE(SUM(total_amount), 0) as revenue
		FROM orders
		WHERE DATE(created_at) BETWEEN ? AND ? AND status = 'paid'
		GROUP BY DATE(created_at)
		ORDER BY date ASC`,
		report.WeekStart, report.WeekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily summary: %w", err)
	}
	if report.DailySummary == nil {
		report.DailySummary = []DailySummary{}
	}

	return report, nil
}

// GetTopProducts returns top selling products for a date range
func (r *Repository) GetTopProducts(startDate, endDate string, limit int) ([]ProductSales, error) {
	if limit <= 0 {
		limit = 10
	}

	var products []ProductSales
	err := r.db.Select(&products, `
		SELECT 
			oi.product_id,
			oi.product_name,
			SUM(oi.quantity) as quantity_sold,
			SUM(oi.subtotal) as revenue
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.id
		WHERE DATE(o.created_at) BETWEEN ? AND ? AND o.status = 'paid'
		GROUP BY oi.product_id, oi.product_name
		ORDER BY quantity_sold DESC
		LIMIT ?`, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top products: %w", err)
	}
	if products == nil {
		products = []ProductSales{}
	}
	return products, nil
}
