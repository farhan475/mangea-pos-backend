package activity_log

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

func (r *Repository) Create(req CreateActivityLogRequest) (*ActivityLog, error) {
	timestamp, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		timestamp = time.Now()
	}

	id := req.ID
	if id == "" {
		id = util.GenerateID()
	}

	log := ActivityLog{
		ID:          id,
		Type:        req.Type,
		Action:      req.Action,
		Description: req.Description,
		EntityID:    req.EntityID,
		EntityType:  req.EntityType,
		Metadata:    req.Metadata,
		UserID:      req.UserID,
		Timestamp:   timestamp,
	}

	query := `
		INSERT INTO activity_logs (id, type, action, description, entity_id, entity_type, metadata, user_id, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			type = VALUES(type),
			action = VALUES(action),
			description = VALUES(description),
			entity_id = VALUES(entity_id),
			entity_type = VALUES(entity_type),
			metadata = VALUES(metadata),
			user_id = VALUES(user_id),
			timestamp = VALUES(timestamp)
	`
	_, err = r.db.Exec(query, log.ID, log.Type, log.Action, log.Description,
		log.EntityID, log.EntityType, log.Metadata, log.UserID, log.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to create activity log: %w", err)
	}

	return &log, nil
}

func (r *Repository) List(logType *string, limit int) ([]ActivityLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	query := "SELECT id, type, action, description, entity_id, entity_type, metadata, user_id, timestamp FROM activity_logs WHERE 1=1"
	args := []interface{}{}

	if logType != nil && *logType != "" {
		query += " AND type = ?"
		args = append(args, *logType)
	}

	query += " ORDER BY timestamp DESC LIMIT ?"
	args = append(args, limit)

	var logs []ActivityLog
	if err := r.db.Select(&logs, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list activity logs: %w", err)
	}

	return logs, nil
}

func (r *Repository) GetByID(id string) (*ActivityLog, error) {
	var log ActivityLog
	err := r.db.Get(&log, "SELECT id, type, action, description, entity_id, entity_type, metadata, user_id, timestamp FROM activity_logs WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get activity log: %w", err)
	}
	return &log, nil
}

func (r *Repository) DeleteOlderThan(days int) error {
	query := "DELETE FROM activity_logs WHERE timestamp < DATE_SUB(NOW(), INTERVAL ? DAY)"
	_, err := r.db.Exec(query, days)
	if err != nil {
		return fmt.Errorf("failed to delete old activity logs: %w", err)
	}
	return nil
}
