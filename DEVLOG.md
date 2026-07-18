# DEVLOG — 2026-07-19

## Session Overview
Batch development session: sorting + 3-kolom stok + mutasi + trace.
Repo: `apotek-clean`, working dir: `/home/jarvis/.dev/apotek-clean`

---

## Changes in This Session

### 1. Sort Produk A-Z ✅
- **File**: `internal/adapters/driving/http/handlers/product_handler.go`
- **Change**: `GetAllProduct` query + `ORDER BY pro.name ASC`
- **Commit**: a8fb500

---

### 2. Three-Column Stock (showcase + warehouse + total) ✅
- **Files**: 15 files changed
- **Concept**: 
  - `products.showcase_stock` = stok etalase (berubah karena transaksi)
  - `products.warehouse_stock` = stok gudang (base stock)
  - `products.stock` = total (computed, read-only display)
- **Logic rule**:
  - Sale / Sale Return / Duplicate Receipt / Opname = `showcase_stock`
  - Purchase / First Stock / Buy Return = `warehouse_stock`
  - Mutasi = transfer antar lokasi (total tetap)
- **Files modified**:
  - `internal/core/entities/master_product_model.go` — ShowcaseStock + WarehouseStock
  - `configs/database_config.go` — AutoMigrate
  - `services/stock_service.go` — all functions → showcase_stock
  - `services/sale_service.go` — check + update showcase_stock
  - `services/sale_handler.go` — restore showcase_stock on delete
  - `services/purchase_service.go` — warehouse_stock
  - `services/purchase_handler.go` — warehouse_stock
  - `services/sale_return_service.go` — showcase_stock
  - `services/sale_return_handler.go` — showcase_stock (fixed raw SQL bug)
  - `services/buy_return_service.go` — warehouse_stock
  - `services/buy_return_handler.go` — warehouse_stock (fixed raw SQL bug)
  - `services/first_stock_service.go` — warehouse_stock
  - `services/duplicate_receipt_service.go` — showcase_stock
  - `services/opname_service.go` — showcase_stock
  - `internal/adapters/driving/http/handlers/product_handler.go` — GetProduct + CmbProdSale include 3 stock fields
- **Commit**: a8fb500

---

### 3. Mutasi Stok ✅
- **Menu name**: "Mutasi Stok"
- **Table**: `stock_mutations`
- **Files created**:
  - `internal/core/entities/stock_mutation_model.go`
  - `services/stock_mutation_service.go`
  - `internal/adapters/driving/http/handlers/stock_mutation_handler.go`
- **Routes**:
  - `POST /api/stock-mutations` — create mutation (roles: admin, operator, superadmin)
  - `GET /api/stock-mutations` — list mutations (roles: admin, operator, cashier, finance, superadmin)
- **Logic**: Atomic transaction — check stock source → decrement source → increment destination
- **Commit**: a8fb500

---

### 4. Product Trace (Riwayat Stok) ✅
- **Menu name**: "Riwayat Stok"
- **Table**: `stock_traces`
- **Files created**:
  - `internal/core/entities/stock_trace_model.go`
  - `services/stock_trace_service.go`
  - `internal/adapters/driving/http/handlers/stock_trace_handler.go`
- **Routes**:
  - `GET /api/product-traces` — trace by product (roles: all authenticated)
  - `GET /api/product-traces/range` — trace by date range (roles: all authenticated)
- **Logged types**: purchase, first_stock, sale, sale_return, buy_return, opname, mutation_in, mutation_out
- **Commit**: a8fb500

---

### 5. Frontend Menu Placement
- Mutasi Stok → Settings > Profile (sub-menu)
- Riwayat Stok → Finance > Defecta (sub-menu)
- **Status**: PENDING (frontend — separate task)

---

## Bugs Fixed During Review
- `stock_trace_service.go`: import `entities` → alias `models`
- `product_handler.go GetProduct`: missing showcase_stock + warehouse_stock
- `product_handler.go CmbProdSale`: missing warehouse_stock

## Deployment
- Commit: a8fb500 ✅ pushed to origin/main
- Binary: `bin/apotek` 55MB ✅ built
- Service: `apotek-clean.service` ✅ restarted (PID 22344)

---

## Previous Session (2026-07-18) — Refactors
- `SubtractProductStock` → `RestoreProductStock` (sale delete + duplicate receipt rollback)
- `NewExcelServices` + `NewPDFService` → `NewExportService` (consolidated factory)
- Export architecture comments added (export_routes.go + handler files)
- Commit `b10595d`

## Previous Session (2026-07-18) — Stock Bugs (PENDING FIX)
- 🔴 `sale_return_handler.go:163`: raw SQL `stock + ?` → FIXED (showcase_stock)
- 🔴 `buy_return_handler.go:178`: raw SQL `stock - ?` → FIXED (warehouse_stock)
- 🟡 `opname_service.go`: lost update risk — no pessimistic lock (MED — defer)
- 🟡 `ZeroProductStock`: qty param ignored (MED — defer)
- 🟢 No batch/expiry tracking (design decision — needs Abi confirmation)
