-- Activity logs table
CREATE TABLE IF NOT EXISTS activity_logs (
    id          VARCHAR(36)   NOT NULL PRIMARY KEY,
    type        VARCHAR(50)   NOT NULL,
    action      VARCHAR(255)  NOT NULL,
    description TEXT          NOT NULL,
    entity_id   VARCHAR(36)   NULL,
    entity_type VARCHAR(50)   NULL,
    metadata    JSON          NULL,
    user_id     VARCHAR(36)   NULL,
    timestamp   DATETIME(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    INDEX idx_activity_logs_type (type),
    INDEX idx_activity_logs_timestamp (timestamp),
    INDEX idx_activity_logs_entity (entity_id, entity_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
