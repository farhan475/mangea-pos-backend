# Mangea POS Backend

A REST API backend for Mangea POS (Point of Sale) system built with Go, Gin, and MySQL.

## Stack

- **Framework**: Gin Web Framework
- **Database**: MySQL 8.0+
- **Database Access**: sqlx with raw SQL
- **Language**: Go 1.25+

## Quick Start

### Prerequisites

- Go 1.25+
- MySQL 8.0+ (local or Docker)
- Docker & Docker Compose (optional, for local MySQL)

### Setup Instructions

#### 1. Clone and install dependencies

```bash
cd /home/farhan/mangea-backend
go mod download
```

#### 2. Start MySQL (choose one method)

**Option A: Using Docker Compose (recommended for dev)**
```bash
docker-compose up -d
sleep 5  # Wait for MySQL to be ready
```

**Option B: Using local MySQL**
- Ensure MySQL is running on `127.0.0.1:3306`
- Create database: `CREATE DATABASE mangea;`
- Create root user if not exists

#### 3. Apply database migrations

```bash
mysql -h 127.0.0.1 -u root -proot -D mangea < migrations/0001_init.up.sql
```

Or using Makefile:
```bash
make docker-up
make migrate-up
```

#### 4. Run the API server

```bash
go run ./cmd/api
```

Or build and run:
```bash
go build -o bin/api ./cmd/api
./bin/api
```

Server starts on `http://localhost:8080`

## API Endpoints

Base URL: `http://localhost:8080/api/v1`

### Categories
- `GET /categories` - List all categories

### Products
- `GET /products` - List all products (query filters: `category_id`, `is_available`)
- `POST /products` - Create new product
- `GET /products/:id` - Get product by ID
- `PUT /products/:id` - Update product
- `DELETE /products/:id` - Delete product

### Tables
- `GET /tables` - List all tables (query filter: `status`)
- `POST /tables` - Create new table
- `GET /tables/:id` - Get table by ID
- `PUT /tables/:id` - Update table
- `PATCH /tables/:id/status` - Update table status
- `DELETE /tables/:id` - Delete table

### Orders
- `GET /orders` - List orders (query filters: `status`, `table_number`, `customer_name`)
- `POST /orders` - Create new order (with items)
- `GET /orders/:id` - Get order by ID (with items)
- `PUT /orders/:id` - Update full order (status, payment info, customer details)
- `PATCH /orders/:id/status` - Update order status only

### Sync (Offline-First)
- `POST /sync/orders` - Batch upsert orders from offline clients

### Dashboard
- `GET /dashboard/metrics` - Get dashboard metrics (new orders, total today, growth, waiting list)
- `GET /dashboard/popular-dishes?limit=5` - Get popular dishes
- `GET /dashboard/out-of-stock` - Get out of stock items

## Request/Response Format

All requests and responses use JSON with snake_case field names.

### Example: Create Product

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "category_id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Nasi Goreng",
    "price": 25000.00,
    "stock": 10,
    "image_url": "https://example.com/nasi-goreng.jpg",
    "is_available": true
  }'
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "category_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Nasi Goreng",
  "price": 25000.00,
  "stock": 10,
  "image_url": "https://example.com/nasi-goreng.jpg",
  "is_available": true,
  "created_at": "2024-08-30T10:00:00Z",
  "updated_at": "2024-08-30T10:00:00Z"
}
```

## Order Status Transitions

Valid status transitions:
- `pending` → `cooking`, `cancelled`
- `cooking` → `ready`, `cancelled`
- `ready` → `paid`, `cancelled`
- `paid` → (terminal)
- `cancelled` → (terminal)

**Example: Update Order Status**
```bash
curl -X PATCH http://localhost:8080/api/v1/orders/550e8400-e29b-41d4-a716-446655440001/status \
  -H "Content-Type: application/json" \
  -d '{"status": "cooking"}'
```

## Database Schema

### Tables
- `categories` - Product categories
- `products` - Menu items with pricing and inventory
- `tables` - Restaurant floor plan / table management
- `orders` - Customer orders
- `order_items` - Line items in each order

All primary keys use UUID (VARCHAR(36)).

## Offline-First Sync

The Flutter client can create orders offline and sync them when connectivity returns.

**POST /sync/orders** accepts an array of orders and performs idempotent upsert:
- Orders with the same ID are de-duplicated
- Newer data (based on `updated_at`) overwrites stale data
- All synced orders receive `sync_status: "synced"` in response

## Environment Variables

Configure via `.env` file:

```
PORT=8080              # API server port
DB_HOST=127.0.0.1      # MySQL host
DB_PORT=3306           # MySQL port
DB_USER=root           # MySQL user
DB_PASS=root           # MySQL password
DB_NAME=mangea         # Database name
```

## Development

### Project Structure

```
mangea-backend/
├── cmd/api/              # Application entry point
├── internal/
│   ├── config/           # Configuration management
│   ├── db/               # Database connection
│   ├── server/           # Router setup
│   ├── apperror/         # Error response handling
│   ├── category/         # Category domain
│   ├── product/          # Product domain
│   ├── table/            # Table domain
│   ├── order/            # Order domain
│   ├── sync/             # Sync handler
│   ├── dashboard/        # Dashboard metrics
│   └── util/             # Utilities (UUID generation, etc)
├── migrations/           # Database migrations
├── Makefile              # Development targets
├── docker-compose.yml    # MySQL container
└── go.mod               # Go module definition
```

### Architecture

Each domain (product, order, table, etc) follows vertical-slice pattern:
- `model.go` - Data structures with db and json tags
- `repository.go` - Database queries using sqlx
- `handler.go` - HTTP handlers and route registration

No interfaces layer unless there's a real reason (keeping code simple).

## Testing

Manual testing via curl or Postman:

```bash
# List all products
curl http://localhost:8080/api/v1/products

# Create a category first
curl -X POST http://localhost:8080/api/v1/categories \
  -H "Content-Type: application/json" \
  -d '{"name": "Makanan"}'

# Create a product
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{"category_id": "...", "name": "Nasi Goreng", "price": 25000, "is_available": true}'

# List dashboard metrics
curl http://localhost:8080/api/v1/dashboard/metrics
```

## Troubleshooting

### Database Connection Failed
- Ensure MySQL is running on the configured host:port
- Check credentials in `.env` file
- Verify database `mangea` exists

### Port Already in Use
- Change PORT in `.env` file or
- Kill existing process: `lsof -ti:8080 | xargs kill -9`

### Migration Errors
- Ensure MySQL is running and accessible
- Check that `mangea` database exists
- Verify migrations are syntactically correct

## Future Enhancements

- [ ] JWT authentication
- [ ] WebSocket for real-time updates
- [ ] Pagination for list endpoints
- [ ] Advanced reporting (daily/weekly/monthly)
- [ ] Printer integration endpoints
- [ ] User management (staff, roles, permissions)

## License

Proprietary - Mangea POS System
