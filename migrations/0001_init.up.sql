-- Categories table
CREATE TABLE IF NOT EXISTS categories (
    id         VARCHAR(36)  NOT NULL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    created_at DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Products table
CREATE TABLE IF NOT EXISTS products (
    id            VARCHAR(36)    NOT NULL PRIMARY KEY,
    category_id   VARCHAR(36)    NOT NULL,
    name          VARCHAR(255)   NOT NULL,
    price         DECIMAL(15,2)  NOT NULL,
    stock         INT            NOT NULL DEFAULT 0,
    image_url     VARCHAR(500)   NULL,
    is_available  BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at    DATETIME(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at    DATETIME(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    CONSTRAINT fk_products_category FOREIGN KEY (category_id) REFERENCES categories(id)
        ON DELETE RESTRICT ON UPDATE CASCADE,
    INDEX idx_products_category_id (category_id),
    INDEX idx_products_availability (is_available, stock)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Tables table
CREATE TABLE IF NOT EXISTS `tables` (
    id               VARCHAR(36)  NOT NULL PRIMARY KEY,
    table_number     VARCHAR(10)  NOT NULL UNIQUE,
    capacity         INT          NOT NULL,
    status           ENUM('available','occupied','reserved') NOT NULL DEFAULT 'available',
    current_order_id VARCHAR(36)  NULL,
    created_at       DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at       DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    INDEX idx_tables_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Orders table
CREATE TABLE IF NOT EXISTS orders (
    id             VARCHAR(36)   NOT NULL PRIMARY KEY,
    user_id        VARCHAR(36)   NULL,
    customer_name  VARCHAR(255)  NULL,
    table_number   VARCHAR(10)   NULL,
    total_amount   DECIMAL(15,2) NOT NULL,
    status         ENUM('pending','cooking','ready','paid','cancelled') NOT NULL DEFAULT 'pending',
    payment_method ENUM('cash','qris','debit') NULL,
    sync_status    ENUM('synced','pending') NOT NULL DEFAULT 'synced',
    created_at     DATETIME(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at     DATETIME(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    INDEX idx_orders_status (status),
    INDEX idx_orders_created_at (created_at),
    INDEX idx_orders_status_created_at (status, created_at),
    INDEX idx_orders_table_number (table_number)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Add FK for tables.current_order_id (after orders table exists)
ALTER TABLE `tables` ADD CONSTRAINT fk_tables_current_order
    FOREIGN KEY (current_order_id) REFERENCES orders(id)
    ON DELETE SET NULL ON UPDATE CASCADE;

-- Order items table
CREATE TABLE IF NOT EXISTS order_items (
    id           VARCHAR(36)   NOT NULL PRIMARY KEY,
    order_id     VARCHAR(36)   NOT NULL,
    product_id   VARCHAR(36)   NOT NULL,
    product_name VARCHAR(255)  NOT NULL,
    price        DECIMAL(15,2) NOT NULL,
    quantity     INT           NOT NULL,
    subtotal     DECIMAL(15,2) NOT NULL,
    CONSTRAINT fk_order_items_order FOREIGN KEY (order_id) REFERENCES orders(id)
        ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_order_items_product FOREIGN KEY (product_id) REFERENCES products(id)
        ON DELETE RESTRICT ON UPDATE CASCADE,
    INDEX idx_order_items_order_id (order_id),
    INDEX idx_order_items_product_id (product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
