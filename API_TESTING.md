# API Testing Guide

Manual testing guide untuk Mangea POS Backend API.

## Prerequisites

- Backend running on `http://localhost:8080`
- MySQL database initialized with migrations
- `curl` command available (or Postman)

## Quick Test Sequence

### 1. Test Database Health

```bash
# Just checking if server is running
curl http://localhost:8080/api/v1/categories
# Response: 200 OK with empty array or categories
```

### 2. Create a Category

```bash
curl -X POST http://localhost:8080/api/v1/categories \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Makanan"
  }'
```

**Response:** 201 Created with category object including `id`, `name`, `created_at`

Save the `id` for next step.

### 3. Create a Product

```bash
CATEGORY_ID="[paste-id-from-step-2]"

curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "category_id": "'$CATEGORY_ID'",
    "name": "Nasi Goreng",
    "price": 25000.00,
    "stock": 10,
    "image_url": "https://example.com/nasi-goreng.jpg",
    "is_available": true
  }'
```

**Response:** 201 Created with product object including `id`, all fields, `created_at`, `updated_at`

Save the product `id`.

### 4. Create a Table

```bash
curl -X POST http://localhost:8080/api/v1/tables \
  -H "Content-Type: application/json" \
  -d '{
    "table_number": "A1",
    "capacity": 4
  }'
```

**Response:** 201 Created with table object. Status defaults to `"available"`.

Save the table `id`.

### 5. Create an Order

```bash
PRODUCT_ID="[paste-product-id-from-step-3]"

curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_name": "John Doe",
    "table_number": "A1",
    "items": [
      {
        "product_id": "'$PRODUCT_ID'",
        "product_name": "Nasi Goreng",
        "price": 25000.00,
        "quantity": 2,
        "subtotal": 50000.00
      }
    ]
  }'
```

**Response:** 201 Created with order object. Status is `"pending"`, `total_amount` is auto-calculated (should be 50000.00).

Save the order `id`.

### 6. List Orders

```bash
curl http://localhost:8080/api/v1/orders
```

**Response:** 200 OK with array of orders (includes your just-created order).

### 7. Update Order Status (PATCH)

```bash
ORDER_ID="[paste-order-id-from-step-5]"

curl -X PATCH http://localhost:8080/api/v1/orders/$ORDER_ID/status \
  -H "Content-Type: application/json" \
  -d '{"status": "cooking"}'
```

**Response:** 200 OK with updated order. Status changed to `"cooking"`.

### 8. Update Full Order (PUT)

```bash
curl -X PUT http://localhost:8080/api/v1/orders/$ORDER_ID \
  -H "Content-Type: application/json" \
  -d '{
    "customer_name": "Jane Doe",
    "table_number": "A1",
    "total_amount": 50000.00,
    "status": "ready",
    "payment_method": "cash"
  }'
```

**Response:** 200 OK with updated order. Name, status, and payment method changed.

### 9. Update Order to Paid (Frees Table)

```bash
curl -X PATCH http://localhost:8080/api/v1/orders/$ORDER_ID/status \
  -H "Content-Type: application/json" \
  -d '{"status": "paid"}'
```

**Response:** 200 OK. Order status is now `"paid"`.

Verify table status changed back to `"available"`:
```bash
curl http://localhost:8080/api/v1/tables
```

### 10. Test Dashboard Metrics

```bash
curl http://localhost:8080/api/v1/dashboard/metrics
```

**Response:** 200 OK
```json
{
  "new_orders_count": 0,
  "total_orders_today": 1,
  "orders_growth_percent": null,
  "waiting_list_count": 0
}
```

### 11. Test Popular Dishes

```bash
curl "http://localhost:8080/api/v1/dashboard/popular-dishes?limit=5"
```

**Response:** 200 OK with array of dishes (your Nasi Goreng x2 quantity = most popular)

### 12. Test Out of Stock

Create a product with `stock=0`:

```bash
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "category_id": "'$CATEGORY_ID'",
    "name": "Rendang",
    "price": 30000.00,
    "stock": 0,
    "is_available": false
  }'
```

Then list out of stock:
```bash
curl http://localhost:8080/api/v1/dashboard/out-of-stock
```

**Response:** 200 OK with products that are unavailable or out of stock.

## Error Cases Testing

### Invalid Status Transition

```bash
# Try invalid transition: paid -> cooking (not allowed)
curl -X PATCH http://localhost:8080/api/v1/orders/$ORDER_ID/status \
  -H "Content-Type: application/json" \
  -d '{"status": "cooking"}'
```

**Response:** 422 Unprocessable Entity
```json
{"error": "cannot transition from paid to cooking"}
```

### Non-existent Resource

```bash
curl http://localhost:8080/api/v1/orders/non-existent-id
```

**Response:** 404 Not Found
```json
{"error": "order not found"}
```

### Bad Request (Missing Required Field)

```bash
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Product without category"
  }'
```

**Response:** 400 Bad Request
```json
{"error": "Key: 'CreateProductRequest.CategoryID' Error:Field validation for 'CategoryID' failed on the 'required' tag"}
```

## Batch Sync Testing (Offline-First)

```bash
curl -X POST http://localhost:8080/api/v1/sync/orders \
  -H "Content-Type: application/json" \
  -d '[
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "customer_name": "Offline User",
      "table_number": "B2",
      "total_amount": 100000.00,
      "status": "pending",
      "sync_status": "pending",
      "created_at": "2024-08-30T10:00:00Z",
      "updated_at": "2024-08-30T10:00:00Z",
      "items": [
        {
          "id": "550e8400-e29b-41d4-a716-446655440002",
          "order_id": "550e8400-e29b-41d4-a716-446655440001",
          "product_id": "'$PRODUCT_ID'",
          "product_name": "Nasi Goreng",
          "price": 50000.00,
          "quantity": 2,
          "subtotal": 100000.00
        }
      ]
    }
  ]'
```

**Response:** 200 OK with synced orders array (sync_status = "synced")

Send same request again - should be idempotent (no duplicate, just updates).

## Test Automation Script

See `test.sh` for automated testing (coming soon).

## Postman Collection

Import the following into Postman for easier testing:

[Postman Collection JSON export coming soon]

## Notes

- All timestamps are in ISO8601 format (RFC3339)
- All monetary values are in IDR (as decimals)
- UUIDs are version 4 strings (36 characters)
- Null fields are represented as `null` in JSON
- Order status transitions are validated on the backend
- Table status auto-updates when orders change
