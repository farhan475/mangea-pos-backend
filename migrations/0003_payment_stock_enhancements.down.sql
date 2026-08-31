-- Rollback Migration 0003: Payment & Stock Management Enhancements

-- 1. Remove index for low stock products
DROP INDEX idx_products_low_stock ON products;

-- 2. Remove low_stock_threshold from products
ALTER TABLE products DROP COLUMN low_stock_threshold;

-- 3. Remove paid_amount and change_amount from orders
ALTER TABLE orders 
DROP COLUMN change_amount,
DROP COLUMN paid_amount;

-- 4. Revert payment_method ENUM to original values
ALTER TABLE orders 
MODIFY COLUMN payment_method ENUM('cash','qris','debit') NULL;
