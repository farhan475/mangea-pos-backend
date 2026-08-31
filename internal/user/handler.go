package user

import (
	"net/http"
	"time"

	"mangea-backend/internal/apperror"
	"mangea-backend/internal/auth"
	"mangea-backend/internal/util"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	repo      *Repository
	jwtSecret string
}

func NewHandler(repo *Repository, jwtSecret string) *Handler {
	return &Handler{repo: repo, jwtSecret: jwtSecret}
}

// Login handles user authentication
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondBadRequest(c, "Invalid request body")
		return
	}

	// Find user by username
	user, err := h.repo.FindByUsername(req.Username)
	if err != nil {
		apperror.RespondInternalServerError(c, "Database error")
		return
	}

	if user == nil {
		apperror.RespondUnauthorized(c, "Invalid username or password")
		return
	}

	// Check if user is active
	if !user.IsActive {
		apperror.RespondUnauthorized(c, "Account is inactive")
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		apperror.RespondUnauthorized(c, "Invalid username or password")
		return
	}

	// Update last login
	_ = h.repo.UpdateLastLogin(user.ID)

	// Generate JWT token
	token, expiresAt, err := auth.GenerateToken(h.jwtSecret, user.ID, user.Role)
	if err != nil {
		apperror.RespondInternalServerError(c, "Failed to generate token")
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		User:      *user,
		Token:     token,
		ExpiresAt: expiresAt.Unix(),
	})
}

// GetCurrentUser returns the authenticated user (from JWT claims)
func (h *Handler) GetCurrentUser(c *gin.Context) {
	userID := auth.UserIDFrom(c)
	if userID == "" {
		apperror.RespondUnauthorized(c, "unauthorized")
		return
	}

	user, err := h.repo.FindByID(userID)
	if err != nil {
		apperror.RespondInternalServerError(c, "Database error")
		return
	}

	if user == nil {
		apperror.RespondNotFound(c, "User not found")
		return
	}

	c.JSON(http.StatusOK, user)
}

// ListUsers retrieves all users
func (h *Handler) ListUsers(c *gin.Context) {
	role := c.Query("role")

	var users []User
	var err error

	if role != "" {
		users, err = h.repo.GetByRole(role)
	} else {
		users, err = h.repo.GetAll()
	}

	if err != nil {
		apperror.RespondInternalServerError(c, "Database error")
		return
	}

	c.JSON(http.StatusOK, users)
}

// GetUserByID retrieves a user by ID
func (h *Handler) GetUserByID(c *gin.Context) {
	id := c.Param("id")

	user, err := h.repo.FindByID(id)
	if err != nil {
		apperror.RespondInternalServerError(c, "Database error")
		return
	}

	if user == nil {
		apperror.RespondNotFound(c, "User not found")
		return
	}

	c.JSON(http.StatusOK, user)
}

// CreateUser creates a new user
func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondBadRequest(c, err.Error())
		return
	}

	// Check if username already exists
	existing, err := h.repo.FindByUsername(req.Username)
	if err != nil {
		apperror.RespondInternalServerError(c, "Database error")
		return
	}
	if existing != nil {
		apperror.RespondConflict(c, "Username already exists")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		apperror.RespondInternalServerError(c, "Failed to hash password")
		return
	}

	// Create user
	user := &User{
		ID:        util.GenerateID(),
		Username:  req.Username,
		Password:  string(hashedPassword),
		Name:      req.Name,
		Role:      req.Role,
		IsActive:  req.IsActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.repo.Create(user); err != nil {
		apperror.RespondInternalServerError(c, "Failed to create user")
		return
	}

	c.JSON(http.StatusCreated, user)
}

// UpdateUser updates an existing user
func (h *Handler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondBadRequest(c, err.Error())
		return
	}

	// Check if user exists
	user, err := h.repo.FindByID(id)
	if err != nil {
		apperror.RespondInternalServerError(c, "Database error")
		return
	}
	if user == nil {
		apperror.RespondNotFound(c, "User not found")
		return
	}

	// Update user
	if err := h.repo.Update(id, req.Name, req.Role, req.IsActive); err != nil {
		apperror.RespondInternalServerError(c, "Failed to update user")
		return
	}

	// Fetch updated user
	updatedUser, _ := h.repo.FindByID(id)
	c.JSON(http.StatusOK, updatedUser)
}

// ChangePassword changes user password
func (h *Handler) ChangePassword(c *gin.Context) {
	id := c.Param("id")

	// Authorization: users may only change their own password, unless admin/owner
	authUserID := auth.UserIDFrom(c)
	authRole := auth.RoleFrom(c)
	if authUserID != id && authRole != string(RoleAdmin) && authRole != string(RoleOwner) {
		apperror.RespondUnauthorized(c, "You can only change your own password")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondBadRequest(c, err.Error())
		return
	}

	// Get user
	user, err := h.repo.FindByID(id)
	if err != nil {
		apperror.RespondInternalServerError(c, "Database error")
		return
	}
	if user == nil {
		apperror.RespondNotFound(c, "User not found")
		return
	}

	// Verify old password (admin/owner may reset without knowing it)
	if authUserID == id {
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword))
		if err != nil {
			apperror.RespondUnauthorized(c, "Invalid old password")
			return
		}
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		apperror.RespondInternalServerError(c, "Failed to hash password")
		return
	}

	// Update password
	if err := h.repo.UpdatePassword(id, string(hashedPassword)); err != nil {
		apperror.RespondInternalServerError(c, "Failed to update password")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// DeleteUser soft deletes a user
func (h *Handler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	// Check if user exists
	user, err := h.repo.FindByID(id)
	if err != nil {
		apperror.RespondInternalServerError(c, "Database error")
		return
	}
	if user == nil {
		apperror.RespondNotFound(c, "User not found")
		return
	}

	// Soft delete
	if err := h.repo.Delete(id); err != nil {
		apperror.RespondInternalServerError(c, "Failed to delete user")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
