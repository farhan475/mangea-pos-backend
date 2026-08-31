package table

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

func (r *Repository) List(status *string) ([]Table, error) {
	query := "SELECT id, table_number, capacity, status, current_order_id, created_at, updated_at FROM `tables` t WHERE 1=1"
	args := []interface{}{}

	if status != nil {
		query += " AND status = ?"
		args = append(args, *status)
	}

	query += " ORDER BY table_number ASC"

	var tables []Table
	if err := r.db.Select(&tables, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}

	return tables, nil
}

func (r *Repository) GetByID(id string) (*Table, error) {
	var table Table
	err := r.db.Get(&table, "SELECT id, table_number, capacity, status, current_order_id, created_at, updated_at FROM `tables` WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get table: %w", err)
	}
	return &table, nil
}

func (r *Repository) Create(req CreateTableRequest) (*Table, error) {
	id := util.GenerateID()
	now := time.Now()
	status := req.Status
	if status == "" {
		status = string(StatusAvailable)
	}

	table := Table{
		ID:          id,
		TableNumber: req.TableNumber,
		Capacity:    req.Capacity,
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	query := "INSERT INTO `tables` (id, table_number, capacity, status, created_at, updated_at) VALUES (:id, :table_number, :capacity, :status, :created_at, :updated_at)"

	_, err := r.db.NamedExec(query, table)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &table, nil
}

func (r *Repository) Update(id string, req UpdateTableRequest) (*Table, error) {
	now := time.Now()

	query := "UPDATE `tables` SET table_number = ?, capacity = ?, updated_at = ? WHERE id = ?"

	result, err := r.db.Exec(query, req.TableNumber, req.Capacity, now, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update table: %w", err)
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

func (r *Repository) UpdateStatus(id string, status string) (*Table, error) {
	// Validate status
	validStatuses := map[string]bool{
		string(StatusAvailable): true,
		string(StatusOccupied):   true,
		string(StatusReserved):   true,
	}

	if !validStatuses[status] {
		return nil, fmt.Errorf("invalid status: %s", status)
	}

	now := time.Now()

	query := "UPDATE `tables` SET status = ?, updated_at = ? WHERE id = ?"

	result, err := r.db.Exec(query, status, now, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update table status: %w", err)
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
	result, err := r.db.Exec("DELETE FROM `tables` WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete table: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("table not found")
	}

	return nil
}
