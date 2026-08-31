package activity_log

import "time"

type ActivityLog struct {
	ID          string            `db:"id" json:"id"`
	Type        string            `db:"type" json:"type"`
	Action      string            `db:"action" json:"action"`
	Description string            `db:"description" json:"description"`
	EntityID    *string           `db:"entity_id" json:"entity_id"`
	EntityType  *string           `db:"entity_type" json:"entity_type"`
	Metadata    *string           `db:"metadata" json:"metadata"`
	UserID      *string           `db:"user_id" json:"user_id"`
	Timestamp   time.Time         `db:"timestamp" json:"timestamp"`
}

type CreateActivityLogRequest struct {
	ID          string  `json:"id" binding:"required"`
	Type        string  `json:"type" binding:"required"`
	Action      string  `json:"action" binding:"required"`
	Description string  `json:"description" binding:"required"`
	EntityID    *string `json:"entity_id"`
	EntityType  *string `json:"entity_type"`
	Metadata    *string `json:"metadata"`
	UserID      *string `json:"user_id"`
	Timestamp   string  `json:"timestamp" binding:"required"`
}

type ListActivityLogRequest struct {
	Type  *string `form:"type"`
	Limit *int    `form:"limit"`
}
