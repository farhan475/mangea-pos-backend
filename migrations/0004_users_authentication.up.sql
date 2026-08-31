-- Migration 0004: User Management & Authentication
-- Description: Create users table with role-based access control

-- 1. Create users table
CREATE TABLE IF NOT EXISTS users (
    id              VARCHAR(36)  NOT NULL PRIMARY KEY,
    username        VARCHAR(50)  NOT NULL UNIQUE,
    password        VARCHAR(255) NOT NULL COMMENT 'Bcrypt hashed password',
    name            VARCHAR(255) NOT NULL,
    role            ENUM('admin','kasir','owner') NOT NULL DEFAULT 'kasir',
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    last_login_at   DATETIME(3)  NULL,
    created_at      DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at      DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    INDEX idx_users_username (username),
    INDEX idx_users_role (role),
    INDEX idx_users_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 2. Insert default users (password: 'admin123', 'kasir123', 'owner123')
INSERT INTO users (id, username, password, name, role, is_active) VALUES
('admin-001', 'admin', '$2a$10$Eeq4763hxb6AouuuMp.30.WwCfSAfMfPgKr67Ll/Td1uQnWpthFrO', 'Administrator', 'admin', TRUE),
('kasir-001', 'kasir', '$2a$10$qWdSXR5uWm53eoK/LDKXwOzgVxzKhd9hnx8QSM5bD0lQU80Mfe/XC', 'Kasir 1', 'kasir', TRUE),
('owner-001', 'owner', '$2a$10$zQr38RNAW9qEEDXuUeofa.IxVMH7v4WuyQnCQqVal55pbho3b9Vne', 'Owner', 'owner', TRUE);

-- 3. Add foreign key to orders.user_id (if not exists)
-- First, update existing orders to have a valid user_id
UPDATE orders SET user_id = 'admin-001' WHERE user_id IS NULL OR user_id = '';

-- Then add the foreign key constraint
ALTER TABLE orders 
ADD CONSTRAINT fk_orders_user 
FOREIGN KEY (user_id) REFERENCES users(id)
ON DELETE SET NULL ON UPDATE CASCADE;

-- 4. Create index for orders by user
CREATE INDEX idx_orders_user_id ON orders(user_id);
