package user

import "time"

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleKasir UserRole = "kasir"
	RoleOwner UserRole = "owner"
)

type User struct {
	ID          string     `db:"id" json:"id"`
	Username    string     `db:"username" json:"username"`
	Password    string     `db:"password" json:"-"` // Never expose password in JSON
	Name        string     `db:"name" json:"name"`
	Role        string     `db:"role" json:"role"`
	IsActive    bool       `db:"is_active" json:"is_active"`
	LastLoginAt *time.Time `db:"last_login_at" json:"last_login_at"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	User      User   `json:"user"`
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=admin kasir owner"`
	IsActive bool   `json:"is_active"`
}

type UpdateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=admin kasir owner"`
	IsActive bool   `json:"is_active"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
