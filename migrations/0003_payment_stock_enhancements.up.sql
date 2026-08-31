-- Migration 0003: Payment & Stock Management Enhancements
-- Description: Add payment details fields, update payment methods, and add stock threshold

-- 1. Update payment_method ENUM to include 'card' and 'ewallet'
ALTER TABLE orders 
MODIFY COLUMN payment_method ENUM('cash','card','ewallet','qris') NULL;

-- 2. Add paid_amount and change_amount for cash payment tracking
ALTER TABLE orders 
ADD COLUMN paid_amount DECIMAL(15,2) NULL AFTER payment_method,
ADD COLUMN change_amount DECIMAL(15,2) NULL AFTER paid_amount;

-- 3. Add low_stock_threshold to products table
ALTER TABLE products 
ADD COLUMN low_stock_threshold INT NOT NULL DEFAULT 10 AFTER stock;

-- 4. Add index for low stock products query optimization
CREATE INDEX idx_products_low_stock ON products(stock, low_stock_threshold);

-- 5. Add comment to clarify payment fields usage
ALTER TABLE orders MODIFY COLUMN paid_amount DECIMAL(15,2) NULL 
COMMENT 'Amount paid by customer (for cash payments)';

ALTER TABLE orders MODIFY COLUMN change_amount DECIMAL(15,2) NULL 
COMMENT 'Change to return to customer (for cash payments)';
