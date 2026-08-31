-- Drop order_items first (FK to orders)
DROP TABLE IF EXISTS order_items;

-- Drop tables table (FK to orders)
ALTER TABLE `tables` DROP FOREIGN KEY fk_tables_current_order;
DROP TABLE IF EXISTS `tables`;

-- Drop orders table
DROP TABLE IF EXISTS orders;

-- Drop products table (FK to categories)
DROP TABLE IF EXISTS products;

-- Drop categories table
DROP TABLE IF EXISTS categories;
