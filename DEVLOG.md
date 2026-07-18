# DEVLOG — 2026-07-19

## Session Overview
Batch development session: sorting + 3-kolom stok + mutasi + trace.
Repo: `apotek-clean`, working dir: `/home/jarvis/.dev/apotek-clean`

---

## Changes in This Session

### 1. Sort Produk A-Z
- **File**: `internal/adapters/driving/http/handlers/product_handler.go`
- **Change**: `GetAllProduct` query + `ORDER BY pro.name ASC`
- **Status**: DONE ✅
- **Commit**: (pending)

---

### 2. Three-Column Stock (showcase + warehouse + total)
- **Files affected**: TBD by sub-agent
- **Concept**: 
  - `products.showcase_stock` = stok yang berubah karena transaksi
  - `products.warehouse_stock` = stok gudang (base)
  - `products.stock` = computed total (read-only, always = showcase + warehouse)
  - Semua transaksi POS HANYA touch `showcase_stock`
  - Purchase/FristStock masuk ke warehouse
  - Sale/CustomerReturn menggerus showcase
  - BuyReturn menggerus warehouse (ke supplier)
  - Opname menggerus showcase
  - Mutasi: pindahkan warehouse → showcase atau sebaliknya
- **Status**: IN PROGRESS (sub-agent)

---

### 3. Mutasi Stok
- **Files affected**: TBD by sub-agent
- **Menu name**: "Mutasi Stok"
- **Table**: `stock_mutations`
- **Concept**: Perpindahan stok antar lokasi (warehouse ↔ showcase)
- **Status**: IN PROGRESS (sub-agent)

---

### 4. Product Trace (Riwayat Stok)
- **Files affected**: TBD by sub-agent
- **Menu name**: "Riwayat Stok"
- **Table**: `stock_traces`
- **Concept**: Log setiap perubahan stok (masuk, mutasi, keluar, opname) per produk
- **Status**: IN PROGRESS (sub-agent)

---

### 5. Frontend Menu Placement
- Mutasi Stok → Settings > Profile (sub-menu)
- Riwayat Stok → Finance > Defecta (sub-menu)
- **Status**: PENDING (frontend — separate task)

---

## Previous Session (2026-07-18) — Refactors
- `SubtractProductStock` → `RestoreProductStock` (sale delete + duplicate receipt rollback)
- `NewExcelServices` + `NewPDFService` → `NewExportService` (consolidated factory)
- Export architecture comments added (export_routes.go + handler files)
- Commit `b10595d`

## Previous Session (2026-07-18) — Audit Findings
- `token_service.go`: KEEP (100 lines, 100+ call sites)
- Export 2-layer architecture: KEEP + clarify (SoC: file-gen vs HTTP layer)

## Previous Session (2026-07-18) — Stock Bugs (PENDING FIX)
- 🔴 `sale_return_handler.go:163`: raw SQL `stock + ?` bypasses stock guard
- 🔴 `buy_return_handler.go:178`: raw SQL `stock - ?` bypasses stock guard
- 🟡 `opname_service.go`: lost update risk — no pessimistic lock
- 🟡 `ZeroProductStock`: qty param ignored
- 🟢 No batch/expiry tracking (design decision — needs Abi confirmation)
