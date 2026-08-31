package user

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// FindByUsername retrieves a user by username
func (r *Repository) FindByUsername(username string) (*User, error) {
	var user User
	query := `SELECT id, username, password, name, role, is_active, last_login_at, created_at, updated_at 
	          FROM users WHERE username = ?`

	err := r.db.Get(&user, query, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByID retrieves a user by ID
func (r *Repository) FindByID(id string) (*User, error) {
	var user User
	query := `SELECT id, username, password, name, role, is_active, last_login_at, created_at, updated_at 
	          FROM users WHERE id = ?`

	err := r.db.Get(&user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetAll retrieves all users
func (r *Repository) GetAll() ([]User, error) {
	var users []User
	query := `SELECT id, username, password, name, role, is_active, last_login_at, created_at, updated_at 
	          FROM users ORDER BY created_at DESC`

	err := r.db.Select(&users, query)
	return users, err
}

// GetByRole retrieves users by role
func (r *Repository) GetByRole(role string) ([]User, error) {
	var users []User
	query := `SELECT id, username, password, name, role, is_active, last_login_at, created_at, updated_at 
	          FROM users WHERE role = ? ORDER BY created_at DESC`

	err := r.db.Select(&users, query, role)
	return users, err
}

// Create creates a new user
func (r *Repository) Create(user *User) error {
	query := `INSERT INTO users (id, username, password, name, role, is_active, created_at, updated_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.Exec(query, user.ID, user.Username, user.Password, user.Name,
		user.Role, user.IsActive, user.CreatedAt, user.UpdatedAt)
	return err
}

// Update updates an existing user
func (r *Repository) Update(id string, name, role string, isActive bool) error {
	query := `UPDATE users SET name = ?, role = ?, is_active = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, name, role, isActive, time.Now(), id)
	return err
}

// UpdatePassword updates user password
func (r *Repository) UpdatePassword(id, hashedPassword string) error {
	query := `UPDATE users SET password = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, hashedPassword, time.Now(), id)
	return err
}

// UpdateLastLogin updates the last login timestamp
func (r *Repository) UpdateLastLogin(id string) error {
	query := `UPDATE users SET last_login_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, time.Now(), id)
	return err
}

// Delete deletes a user (soft delete by setting is_active to false)
func (r *Repository) Delete(id string) error {
	query := `UPDATE users SET is_active = false, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, time.Now(), id)
	return err
}

// HardDelete permanently deletes a user
func (r *Repository) HardDelete(id string) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}
