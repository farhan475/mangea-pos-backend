package order

import (
	"database/sql"
	"fmt"
	"strings"
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

var allowedTransitions = map[string][]string{
	string(StatusPending):  {string(StatusCooking), string(StatusCanceled)},
	string(StatusCooking):  {string(StatusReady), string(StatusCanceled)},
	string(StatusReady):    {string(StatusPaid), string(StatusCanceled)},
	string(StatusPaid):     {},
	string(StatusCanceled): {},
}

func CanTransition(from, to string) bool {
	if to == "" {
		return false
	}
	for _, s := range allowedTransitions[from] {
		if s == to {
			return true
		}
	}
	return false
}

func (r *Repository) List(status *string, tableNumber *string, customerName *string, dateFilter *string) ([]Order, error) {
	query := "SELECT id, user_id, customer_name, table_number, total_amount, status, payment_method, paid_amount, change_amount, sync_status, created_at, updated_at FROM orders WHERE 1=1"
	args := []interface{}{}

	if status != nil && *status != "" {
		query += " AND status = ?"
		args = append(args, *status)
	}

	if tableNumber != nil && *tableNumber != "" {
		query += " AND table_number = ?"
		args = append(args, *tableNumber)
	}

	if customerName != nil && *customerName != "" {
		query += " AND customer_name LIKE ?"
		args = append(args, "%"+*customerName+"%")
	}

	if dateFilter != nil {
		switch *dateFilter {
		case "today":
			query += " AND DATE(created_at) = CURDATE()"
		case "yesterday":
			query += " AND DATE(created_at) = DATE_SUB(CURDATE(), INTERVAL 1 DAY)"
		}
	}

	query += " ORDER BY created_at DESC"

	var orders []Order
	if err := r.db.Select(&orders, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	// Load items for each order
	for i := range orders {
		items, err := r.getOrderItems(orders[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load order items: %w", err)
		}
		orders[i].Items = items
	}

	return orders, nil
}

func (r *Repository) GetByID(id string) (*Order, error) {
	var order Order
	err := r.db.Get(&order, "SELECT id, user_id, customer_name, table_number, total_amount, status, payment_method, paid_amount, change_amount, sync_status, created_at, updated_at FROM orders WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	items, err := r.getOrderItems(order.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load order items: %w", err)
	}
	order.Items = items

	return &order, nil
}

func (r *Repository) getOrderItems(orderID string) ([]OrderItem, error) {
	var items []OrderItem
	err := r.db.Select(&items, "SELECT id, order_id, product_id, product_name, price, quantity, subtotal FROM order_items WHERE order_id = ? ORDER BY id", orderID)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// productInfo holds the server-verified product data for an order item.
type productInfo struct {
	ID    string
	Name  string
	Price float64
	Stock int
}

// resolveProducts looks up all products for the given items inside the transaction.
func resolveProducts(tx *sqlx.Tx, req CreateOrderRequest) (map[string]productInfo, error) {
	products := make(map[string]productInfo, len(req.Items))
	for _, item := range req.Items {
		if _, seen := products[item.ProductID]; seen {
			continue
		}
		var p productInfo
		err := tx.Get(&p, "SELECT id, name, price, stock FROM products WHERE id = ?", item.ProductID)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found: %s", item.ProductID)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to load product %s: %w", item.ProductID, err)
		}
		products[item.ProductID] = p
	}
	return products, nil
}

// validateAndReduceStock checks availability and atomically decrements stock.
func validateAndReduceStock(tx *sqlx.Tx, items []CreateOrderItemRequest, products map[string]productInfo) error {
	// Aggregate quantities per product first
	quantities := make(map[string]int, len(items))
	for _, item := range items {
		quantities[item.ProductID] += item.Quantity
	}

	for productID, qty := range quantities {
		p := products[productID]

		result, err := tx.Exec(
			"UPDATE products SET stock = stock - ? WHERE id = ? AND stock >= ?",
			qty, productID, qty,
		)
		if err != nil {
			return fmt.Errorf("failed to reduce stock for product %s: %w", productID, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}
		if rowsAffected == 0 {
			return fmt.Errorf("insufficient stock for product %s (%s): requested %d, available %d",
				p.Name, productID, qty, p.Stock)
		}
	}
	return nil
}

func (r *Repository) Create(req CreateOrderRequest) (*Order, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Resolve authoritative product data from DB (never trust client prices)
	products, err := resolveProducts(tx, req)
	if err != nil {
		return nil, err
	}

	// Calculate total amount server-side from DB prices
	totalAmount := 0.0
	for _, item := range req.Items {
		p := products[item.ProductID]
		totalAmount += p.Price * float64(item.Quantity)
	}

	// Reduce stock atomically (fails if insufficient)
	if err := validateAndReduceStock(tx, req.Items, products); err != nil {
		return nil, err
	}

	// Use the client-provided ID when present so local and server records stay
	// aligned (prevents duplicate orders after offline sync). Otherwise generate one.
	id := req.ID
	if id == "" {
		id = util.GenerateID()
	}
	now := time.Now()

	order := Order{
		ID:           id,
		UserID:       req.UserID,
		CustomerName: req.CustomerName,
		TableNumber:  req.TableNumber,
		TotalAmount:  totalAmount,
		Status:       string(StatusPending),
		SyncStatus:   string(SyncStatusSynced),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Insert order
	orderQuery := `
		INSERT INTO orders (id, user_id, customer_name, table_number, total_amount, status, payment_method, sync_status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = tx.Exec(orderQuery, order.ID, order.UserID, order.CustomerName, order.TableNumber, order.TotalAmount, order.Status, order.PaymentMethod, order.SyncStatus, order.CreatedAt, order.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}

	// Insert order items using server-verified product data
	for _, item := range req.Items {
		itemID := util.GenerateID()
		if item.ID != "" {
			itemID = item.ID
		}

		p := products[item.ProductID]
		subtotal := p.Price * float64(item.Quantity)

		itemQuery := `
			INSERT INTO order_items (id, order_id, product_id, product_name, price, quantity, subtotal)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`
		_, err = tx.Exec(itemQuery, itemID, id, item.ProductID, p.Name, p.Price, item.Quantity, subtotal)
		if err != nil {
			return nil, fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	// If table number is provided, update table status
	if order.TableNumber != nil {
		tableQuery := "UPDATE `tables` SET status = 'occupied', current_order_id = ?, updated_at = ? WHERE table_number = ?"
		_, err = tx.Exec(tableQuery, order.ID, now, *order.TableNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to update table: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload order with items
	return r.GetByID(id)
}

func (r *Repository) Update(id string, req UpdateOrderRequest) (*Order, error) {
	// Get current order
	current, err := r.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get current order: %w", err)
	}

	if current == nil {
		return nil, nil
	}

	// If status is changing, validate transition
	if req.Status != "" && req.Status != current.Status {
		if !CanTransition(current.Status, req.Status) {
			return nil, fmt.Errorf("cannot transition from %s to %s", current.Status, req.Status)
		}
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()

	// Prepare status for update
	status := req.Status
	if status == "" {
		status = current.Status
	}

	// Update order
	updateQuery := `UPDATE orders SET user_id = ?, customer_name = ?, table_number = ?, total_amount = ?, status = ?, payment_method = ?, updated_at = ? WHERE id = ?`
	_, err = tx.Exec(updateQuery, req.UserID, req.CustomerName, req.TableNumber, req.TotalAmount, status, req.PaymentMethod, now, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	// If transitioning to paid or cancelled, free the table
	if status != current.Status {
		if status == string(StatusPaid) || status == string(StatusCanceled) {
			if current.TableNumber != nil {
				tableQuery := "UPDATE `tables` SET status = 'available', current_order_id = NULL, updated_at = ? WHERE table_number = ?"
				_, err = tx.Exec(tableQuery, now, *current.TableNumber)
				if err != nil {
					return nil, fmt.Errorf("failed to update table: %w", err)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetByID(id)
}

func (r *Repository) UpdateStatus(id string, newStatus string) (*Order, error) {
	// Validate transition first (read-only check)
	current, err := r.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get current order: %w", err)
	}

	if current == nil {
		return nil, nil
	}

	if !CanTransition(current.Status, newStatus) {
		return nil, fmt.Errorf("cannot transition from %s to %s", current.Status, newStatus)
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()

	// Atomic transition guard: only update if status is still what we validated
	// (prevents race conditions from concurrent status changes)
	updateQuery := `UPDATE orders SET status = ?, updated_at = ? WHERE id = ? AND status = ?`
	result, err := tx.Exec(updateQuery, newStatus, now, id, current.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("order status changed concurrently, please retry")
	}

	// If transitioning to paid or cancelled, free the table
	if newStatus == string(StatusPaid) || newStatus == string(StatusCanceled) {
		if current.TableNumber != nil {
			tableQuery := "UPDATE `tables` SET status = 'available', current_order_id = NULL, updated_at = ? WHERE table_number = ?"
			_, err = tx.Exec(tableQuery, now, *current.TableNumber)
			if err != nil {
				return nil, fmt.Errorf("failed to update table: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetByID(id)
}

// Pay records a payment for an order and marks it as paid.
// Payment data (paid_amount, change_amount) is stored in the orders row.
func (r *Repository) Pay(id string, req PaymentRequest) (*Order, error) {
	current, err := r.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get current order: %w", err)
	}
	if current == nil {
		return nil, nil
	}

	if current.Status == string(StatusPaid) {
		return nil, fmt.Errorf("order is already paid")
	}
	if current.Status != string(StatusReady) && current.Status != string(StatusPending) && current.Status != string(StatusCooking) {
		return nil, fmt.Errorf("cannot pay order with status %s", current.Status)
	}

	if req.PaymentMethod == "cash" && req.PaidAmount < current.TotalAmount {
		return nil, fmt.Errorf("paid amount %.2f is less than total %.2f", req.PaidAmount, current.TotalAmount)
	}

	change := 0.0
	if req.PaymentMethod == "cash" {
		change = req.PaidAmount - current.TotalAmount
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()

	// Persist payment fields AND status atomically
	updateQuery := `
		UPDATE orders
		SET status = ?, payment_method = ?, paid_amount = ?, change_amount = ?, updated_at = ?
		WHERE id = ? AND status = ?
	`
	result, err := tx.Exec(updateQuery, string(StatusPaid), req.PaymentMethod, req.PaidAmount, change, now, id, current.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to record payment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("order status changed concurrently, please retry")
	}

	// Free the table if the order was attached to one
	if current.TableNumber != nil {
		tableQuery := "UPDATE `tables` SET status = 'available', current_order_id = NULL, updated_at = ? WHERE table_number = ?"
		if _, err := tx.Exec(tableQuery, now, *current.TableNumber); err != nil {
			return nil, fmt.Errorf("failed to update table: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetByID(id)
}

func (r *Repository) Delete(id string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get order first to free the table
	current, err := r.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}
	if current == nil {
		return fmt.Errorf("order not found")
	}

	// Free the table if occupied
	if current.TableNumber != nil {
		now := time.Now()
		tableQuery := "UPDATE `tables` SET status = 'available', current_order_id = NULL, updated_at = ? WHERE table_number = ? AND current_order_id = ?"
		_, err = tx.Exec(tableQuery, now, *current.TableNumber, id)
		if err != nil {
			return fmt.Errorf("failed to free table: %w", err)
		}
	}

	// Delete order items first (cascade)
	_, err = tx.Exec("DELETE FROM order_items WHERE order_id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete order items: %w", err)
	}

	// Delete order
	result, err := tx.Exec("DELETE FROM orders WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("order not found")
	}

	return tx.Commit()
}

func (r *Repository) UpsertOrders(orders []Order) ([]Order, error) {
	var result []Order

	for _, order := range orders {
		tx, err := r.db.Beginx()
		if err != nil {
			return nil, fmt.Errorf("failed to begin transaction: %w", err)
		}

		now := time.Now()

		// Check if order exists
		var existingUpdatedAt time.Time
		checkQuery := `SELECT updated_at FROM orders WHERE id = ? LIMIT 1`
		err = tx.Get(&existingUpdatedAt, checkQuery, order.ID)

		if err == nil {
			// Order exists, skip if incoming is not strictly newer (prevents stale payloads
			// from wiping newer server data)
			if !order.UpdatedAt.After(existingUpdatedAt) {
				tx.Rollback()
				continue
			}
			// Update it
			updateQuery := `
				UPDATE orders SET user_id = ?, customer_name = ?, table_number = ?, total_amount = ?, status = ?, payment_method = ?, sync_status = 'synced', updated_at = ?
				WHERE id = ?
			`
			_, err = tx.Exec(updateQuery, order.UserID, order.CustomerName, order.TableNumber, order.TotalAmount, order.Status, order.PaymentMethod, now, order.ID)
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to update order %s: %w", order.ID, err)
			}
		} else if err == sql.ErrNoRows {
			// New order, insert it
			insertQuery := `
				INSERT INTO orders (id, user_id, customer_name, table_number, total_amount, status, payment_method, sync_status, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, 'synced', ?, ?)
			`
			_, err = tx.Exec(insertQuery, order.ID, order.UserID, order.CustomerName, order.TableNumber, order.TotalAmount, order.Status, order.PaymentMethod, order.CreatedAt, now)
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to insert order %s: %w", order.ID, err)
			}
		} else {
			tx.Rollback()
			return nil, fmt.Errorf("failed to check order %s: %w", order.ID, err)
		}

		// Handle order items
		if len(order.Items) > 0 {
			// Delete existing items not in the incoming list (parameterized, no string concat)
			params := make([]interface{}, 0, len(order.Items)+1)
			placeholders := make([]string, 0, len(order.Items))
			params = append(params, order.ID)
			for _, item := range order.Items {
				placeholders = append(placeholders, "?")
				params = append(params, item.ID)
			}
			deleteQuery := fmt.Sprintf(`DELETE FROM order_items WHERE order_id = ? AND id NOT IN (%s)`, strings.Join(placeholders, ","))
			if _, err := tx.Exec(deleteQuery, params...); err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to delete stale items for order %s: %w", order.ID, err)
			}

			// Upsert items; use server-side order.ID (never trust client order_id)
			for _, item := range order.Items {
				itemQuery := `
					INSERT INTO order_items (id, order_id, product_id, product_name, price, quantity, subtotal)
					VALUES (?, ?, ?, ?, ?, ?, ?)
					ON DUPLICATE KEY UPDATE product_id = VALUES(product_id), product_name = VALUES(product_name), price = VALUES(price), quantity = VALUES(quantity), subtotal = VALUES(subtotal)
				`
				if _, err := tx.Exec(itemQuery, item.ID, order.ID, item.ProductID, item.ProductName, item.Price, item.Quantity, item.Subtotal); err != nil {
					tx.Rollback()
					return nil, fmt.Errorf("failed to upsert item %s for order %s: %w", item.ID, order.ID, err)
				}
			}
		}

		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("failed to commit order %s: %w", order.ID, err)
		}

		// Re-read the synced order
		syncedOrder, err := r.GetByID(order.ID)
		if err == nil && syncedOrder != nil {
			result = append(result, *syncedOrder)
		}
	}

	// Surface per-order validation problems without failing the whole batch
	return result, nil
}
