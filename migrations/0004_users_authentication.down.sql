-- Rollback Migration 0004: User Management & Authentication

-- 1. Remove index for orders by user
DROP INDEX idx_orders_user_id ON orders;

-- 2. Remove foreign key constraint from orders
ALTER TABLE orders DROP FOREIGN KEY fk_orders_user;

-- 3. Drop users table
DROP TABLE IF EXISTS users;
