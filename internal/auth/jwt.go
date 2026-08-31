package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	ClaimsUserIDKey = "user_id"
	ClaimsRoleKey   = "role"

	ContextUserIDKey = "auth_user_id"
	ContextRoleKey   = "auth_role"

	defaultTokenTTL = 12 * time.Hour
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken creates a signed JWT for the given user.
func GenerateToken(secret, userID, role string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(defaultTokenTTL)

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return signed, expiresAt, nil
}

// ValidateToken parses and validates a signed JWT, returning its claims.
func ValidateToken(secret, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// AuthMiddleware requires a valid Bearer token and stores user info in context.
func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			abortUnauthorized(c, "missing authorization header")
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			abortUnauthorized(c, "authorization header must be in Bearer <token> format")
			return
		}

		claims, err := ValidateToken(secret, strings.TrimSpace(parts[1]))
		if err != nil {
			abortUnauthorized(c, "invalid or expired token")
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextRoleKey, claims.Role)

		c.Next()
	}
}

// RequireRole restricts access to users with one of the given roles.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextRoleKey)
		if !exists {
			abortUnauthorized(c, "unauthorized")
			return
		}

		roleStr, _ := role.(string)
		for _, r := range roles {
			if roleStr == r {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(403, gin.H{"error": "insufficient permissions"})
	}
}

// UserIDFrom returns the authenticated user ID from the request context.
func UserIDFrom(c *gin.Context) string {
	if v, ok := c.Get(ContextUserIDKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// RoleFrom returns the authenticated user role from the request context.
func RoleFrom(c *gin.Context) string {
	if v, ok := c.Get(ContextRoleKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// WithUser injects auth context values into a background context (for goroutines).
func WithUser(ctx context.Context, userID, role string) context.Context {
	ctx = context.WithValue(ctx, ContextUserIDKey, userID)
	return context.WithValue(ctx, ContextRoleKey, role)
}

func abortUnauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(401, gin.H{"error": message})
}
