package product

import (
	"database/sql"
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

func (r *Repository) List(categoryID *string, isAvailable *bool) ([]Product, error) {
	query := "SELECT id, category_id, name, price, stock, low_stock_threshold, image_url, is_available, created_at, updated_at FROM products WHERE 1=1"
	args := []interface{}{}

	if categoryID != nil {
		query += " AND category_id = ?"
		args = append(args, *categoryID)
	}

	if isAvailable != nil {
		query += " AND is_available = ?"
		args = append(args, *isAvailable)
	}

	query += " ORDER BY created_at DESC"

	var products []Product
	if err := r.db.Select(&products, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	return products, nil
}

func (r *Repository) GetByID(id string) (*Product, error) {
	var product Product
	err := r.db.Get(&product, "SELECT id, category_id, name, price, stock, low_stock_threshold, image_url, is_available, created_at, updated_at FROM products WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return &product, nil
}

func (r *Repository) Create(req CreateProductRequest) (*Product, error) {
	id := util.GenerateID()
	now := time.Now()

	product := Product{
		ID:                id,
		CategoryID:        req.CategoryID,
		Name:              req.Name,
		Price:             req.Price,
		Stock:             req.Stock,
		LowStockThreshold: req.LowStockThreshold,
		ImageURL:          req.ImageURL,
		IsAvailable:       req.IsAvailable,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	query := `
		INSERT INTO products (id, category_id, name, price, stock, low_stock_threshold, image_url, is_available, created_at, updated_at)
		VALUES (:id, :category_id, :name, :price, :stock, :low_stock_threshold, :image_url, :is_available, :created_at, :updated_at)
	`

	_, err := r.db.NamedExec(query, product)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return &product, nil
}

func (r *Repository) Update(id string, req UpdateProductRequest) (*Product, error) {
	now := time.Now()

	query := `
		UPDATE products
		SET category_id = ?, name = ?, price = ?, stock = ?, low_stock_threshold = ?, image_url = ?, is_available = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(query, req.CategoryID, req.Name, req.Price, req.Stock, req.LowStockThreshold, req.ImageURL, req.IsAvailable, now, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, nil
	}

	return r.GetByID(id)
}

func (r *Repository) Delete(id string) error {
	result, err := r.db.Exec("DELETE FROM products WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// Stock management methods

// GetLowStockProducts returns products where stock <= low_stock_threshold
func (r *Repository) GetLowStockProducts() ([]Product, error) {
	query := `SELECT id, category_id, name, price, stock, low_stock_threshold, image_url, is_available, created_at, updated_at 
	          FROM products 
	          WHERE stock <= low_stock_threshold AND stock > 0 
	          ORDER BY stock ASC`

	var products []Product
	if err := r.db.Select(&products, query); err != nil {
		return nil, fmt.Errorf("failed to get low stock products: %w", err)
	}
	return products, nil
}

// GetOutOfStockProducts returns products where stock = 0
func (r *Repository) GetOutOfStockProducts() ([]Product, error) {
	query := `SELECT id, category_id, name, price, stock, low_stock_threshold, image_url, is_available, created_at, updated_at 
	          FROM products 
	          WHERE stock = 0 
	          ORDER BY name ASC`

	var products []Product
	if err := r.db.Select(&products, query); err != nil {
		return nil, fmt.Errorf("failed to get out of stock products: %w", err)
	}
	return products, nil
}

// UpdateStock updates the stock quantity for a product
func (r *Repository) UpdateStock(id string, stock int) error {
	query := "UPDATE products SET stock = ?, updated_at = ? WHERE id = ?"
	if _, err := r.db.Exec(query, stock, time.Now(), id); err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}
	return nil
}

// AddStockAtomically atomically increments stock by the given quantity.
func (r *Repository) AddStockAtomically(id string, quantity int) (int, error) {
	query := "UPDATE products SET stock = stock + ?, updated_at = ? WHERE id = ?"
	result, err := r.db.Exec(query, quantity, time.Now(), id)
	if err != nil {
		return 0, fmt.Errorf("failed to add stock: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return 0, fmt.Errorf("product not found")
	}

	var p Product
	if err := r.db.Get(&p, "SELECT stock FROM products WHERE id = ?", id); err != nil {
		return 0, fmt.Errorf("failed to read updated stock: %w", err)
	}
	return p.Stock, nil
}

// ReduceStockAtomically atomically decrements stock, failing if it would go negative.
func (r *Repository) ReduceStockAtomically(id string, quantity int) (int, error) {
	query := "UPDATE products SET stock = stock - ?, updated_at = ? WHERE id = ? AND stock >= ?"
	result, err := r.db.Exec(query, quantity, time.Now(), id, quantity)
	if err != nil {
		return 0, fmt.Errorf("failed to reduce stock: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return 0, fmt.Errorf("insufficient stock or product not found")
	}

	var p Product
	if err := r.db.Get(&p, "SELECT stock FROM products WHERE id = ?", id); err != nil {
		return 0, fmt.Errorf("failed to read updated stock: %w", err)
	}
	return p.Stock, nil
}

// GetStockStatistics returns aggregated stock statistics
func (r *Repository) GetStockStatistics() (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_products,
			SUM(CASE WHEN stock > low_stock_threshold THEN 1 ELSE 0 END) as in_stock,
			SUM(CASE WHEN stock <= low_stock_threshold AND stock > 0 THEN 1 ELSE 0 END) as low_stock,
			SUM(CASE WHEN stock = 0 THEN 1 ELSE 0 END) as out_of_stock,
			SUM(price * stock) as total_stock_value
		FROM products
	`

	var result struct {
		TotalProducts   int     `db:"total_products"`
		InStock         int     `db:"in_stock"`
		LowStock        int     `db:"low_stock"`
		OutOfStock      int     `db:"out_of_stock"`
		TotalStockValue float64 `db:"total_stock_value"`
	}

	if err := r.db.Get(&result, query); err != nil {
		return nil, fmt.Errorf("failed to get stock statistics: %w", err)
	}

	stats := map[string]interface{}{
		"total_products":    result.TotalProducts,
		"in_stock":          result.InStock,
		"low_stock":         result.LowStock,
		"out_of_stock":      result.OutOfStock,
		"total_stock_value": result.TotalStockValue,
	}

	return stats, nil
}
