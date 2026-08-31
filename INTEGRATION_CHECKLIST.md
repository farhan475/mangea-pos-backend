# Integration Testing Checklist

Checklist lengkap untuk test backend + frontend Mangea POS.

## Prerequisites Checklist

- [ ] MySQL running (local or Docker)
- [ ] Database `mangea` created
- [ ] Migrations applied via `mysql < migrations/0001_init.up.sql`
- [ ] Backend built: `go build -o bin/api ./cmd/api`
- [ ] Backend running: `go run ./cmd/api` or `./bin/api`
- [ ] Server listening on `http://localhost:8080/api/v1`
- [ ] Flutter emulator/device running with network access to backend
- [ ] Flutter app base URL pointing to `http://localhost:8080/api/v1`

## Backend Standalone Testing (No Frontend)

Use `curl` commands from `API_TESTING.md`

### Basic CRUD Operations
- [ ] **Categories:** POST → GET → List
- [ ] **Products:** POST → GET → PUT → DELETE
- [ ] **Tables:** POST → GET → PATCH status → DELETE
- [ ] **Orders:** POST → GET → PATCH status → PUT full update

### Order Workflow
- [ ] Create order with multiple items
- [ ] Verify `total_amount` auto-calculated from items
- [ ] Verify table status changes to `occupied` when order created
- [ ] Update order status: pending → cooking
- [ ] Update order status: cooking → ready
- [ ] Update order status: ready → paid
- [ ] Verify table status changes back to `available` when paid
- [ ] Verify cannot transition from `paid` to `cooking` (422 error)

### Sync & Offline-First
- [ ] Send batch sync order request (POST /sync/orders)
- [ ] Resend same sync request (verify idempotent - no duplicates)
- [ ] Verify `sync_status` changes to `"synced"` in response

### Dashboard Features
- [ ] Create 3 orders to populate metrics
- [ ] GET /dashboard/metrics returns correct counts
- [ ] GET /dashboard/popular-dishes shows items sorted by quantity
- [ ] GET /dashboard/out-of-stock shows unavailable products
- [ ] GET /dashboard/popular-dishes?limit=3 respects limit parameter

## Frontend Integration Testing

### Pre-Integration Setup
- [ ] Flutter app builds successfully (`flutter build`)
- [ ] No compilation errors in lib/data/remote/
- [ ] DioClient configured for `http://localhost:8080/api/v1`

### User Flow: Create Order
- [ ] Open app, navigate to POS screen
- [ ] Browse products (GET /products working)
- [ ] Add items to cart
- [ ] Select table (A1, B2, etc)
- [ ] Submit order
  - [ ] POST /orders sends correct structure
  - [ ] Backend returns 201 Created
  - [ ] Frontend receives order with `id` and `total_amount`
  - [ ] Order appears in Dashboard order list

### User Flow: Update Order Status
- [ ] From Dashboard, find pending order
- [ ] Try to update status to "cooking"
  - [ ] PATCH /orders/:id/status sent with {"status": "cooking"}
  - [ ] Backend returns 200 with updated order
  - [ ] UI updates to show new status
  - [ ] Status badge color changes (Yellow for cooking)
- [ ] Update to "ready" (Green badge)
- [ ] Update to "paid"
  - [ ] Verify table status shown as "available" again
  - [ ] Table freed for next customers

### User Flow: Offline Order Creation
- [ ] Disable network (airplane mode or disconnect WiFi)
- [ ] Create order in app (saves to local Hive DB)
- [ ] Verify order shown as "Pending Sync" in UI
- [ ] Re-enable network
- [ ] Sync Manager auto-syncs (or click sync button)
  - [ ] POST /sync/orders sent to backend
  - [ ] Order synced, badge changes to normal
  - [ ] Verify order persisted on backend

### Dashboard Integration
- [ ] Dashboard loads (GET /dashboard/metrics)
- [ ] Metrics display: New Orders, Total Orders, Growth %, Waiting List
- [ ] Popular Dishes widget shows top selling items
- [ ] Out of Stock section shows unavailable items
- [ ] All counts update after order state changes

### Error Handling in Frontend
- [ ] Try creating order without selecting product (validation)
- [ ] Try invalid status transition (expect error message)
- [ ] Try creating order with zero items (validation)
- [ ] Network error handling (simulate offline, retry)

## Data Format Verification

### JSON Serialization Round-Trip
- [ ] Create product with image_url (nullable)
  - [ ] Send: `{"image_url": "https://example.com/img.jpg"}`
  - [ ] Receive: `"image_url": "https://example.com/img.jpg"`
  - [ ] Store locally, resend in sync
  - [ ] Backend receives intact

- [ ] Create order with null customer_name
  - [ ] Send: `{"customer_name": null}` or omit field
  - [ ] Backend accepts
  - [ ] Receive: `"customer_name": null`

### Datetime Format
- [ ] Create order, check created_at field
  - [ ] Format should be ISO8601: `2024-08-30T14:40:00Z`
  - [ ] Flutter parses correctly with `DateTime.parse()`
  - [ ] No timezone conversion issues

### Enum/Status Strings
- [ ] OrderStatus values: pending, cooking, ready, paid, cancelled
  - [ ] UI displays correct status name
  - [ ] Status transitions validate correctly
  
- [ ] TableStatus: available, occupied, reserved
  - [ ] Table displays correct status badge
  
- [ ] SyncStatus: pending, synced
  - [ ] "Pending Sync" badge shows for pending orders
  - [ ] Badge disappears after sync

## Performance & Load Testing

- [ ] List 100+ products → UI responsive
- [ ] List 100+ orders → UI responsive  
- [ ] Create order with 20 items → calculated total correct
- [ ] Batch sync 50 orders → idempotent, completes successfully
- [ ] Dashboard metrics with 1000+ orders → fast calculation

## Stress Testing

- [ ] Create order on one device, update on another → last-write-wins
- [ ] Rapid status transitions (pending → cooking → ready → paid) → all valid
- [ ] Rapid sync (same data multiple times) → idempotent
- [ ] Network interruption during sync → retry works

## Final Sign-Off

- [ ] All basic CRUD operations work
- [ ] Order workflow (create → cooking → ready → paid → table freed) complete
- [ ] Offline sync mechanism works (create offline, sync online)
- [ ] Dashboard metrics accurate
- [ ] Error handling graceful (no crashes)
- [ ] Data integrity verified (no data loss, no duplicates)
- [ ] JSON serialization round-trips successful
- [ ] Frontend UI reflects backend state correctly

---

## Test Data Seed Script (Optional)

```bash
# Create category
curl -X POST http://localhost:8080/api/v1/categories \
  -H "Content-Type: application/json" \
  -d '{"name": "Makanan"}'

# Create 5 products
# Create 5 tables
# Create 3 sample orders
# Run dashboard queries to populate metrics

# See API_TESTING.md for detailed examples
```

## Troubleshooting

### Backend Connection Refused
- Verify backend running: `ps aux | grep api`
- Verify port 8080: `lsof -i :8080`
- Check firewall: `sudo iptables -L` or check settings

### Database Connection Error
- Verify MySQL running: `ps aux | grep mysqld`
- Verify database exists: `mysql -u root -p -e "SHOW DATABASES;"`
- Verify migrations applied: `mysql -u root -p mangea -e "SHOW TABLES;"`

### Frontend Cannot Connect
- Check Flutter network permission (AndroidManifest.xml, Info.plist)
- Verify emulator can reach host: `adb shell ping 192.168.x.x`
- Check DioClient base URL in code

### JSON Parsing Error
- Enable Dio logging in Flutter: `DioClient` should log requests/responses
- Check server response format in backend logs
- Verify entity fromJson() methods handle all fields

### Order Status Transition Error
- Verify current order status before trying transition
- Check validation rules: `pending → cooking/cancelled`, `cooking → ready/cancelled`, etc
- Invalid transitions return 422 Unprocessable Entity

---

**When all checkboxes are complete, backend + frontend integration is ready for production! ✅**
