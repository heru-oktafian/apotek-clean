# API_NORMALIZATION_PLAN.md

## Tujuan
Menormalkan surface API new repo agar tetap ramah untuk frontend lama, lebih rapi secara kontrak, dan lebih jelas secara hierarki, tanpa memutus kompatibilitas yang penting.

## Batas Program
### Dikecualikan dari normalisasi route
- Seluruh area `opname`
- Seluruh area `opname-item`

### Prinsip umum
- Utamakan perubahan **non-breaking**
- Tambah alias kompatibilitas bila perlu
- Jangan ganti kontrak besar tanpa alasan kuat
- Semua perubahan harus didokumentasikan old vs new
- Semua perubahan idealnya diikuti build + smoke test

## Kategori Status
- `pending` = belum disentuh
- `in_progress` = sedang dianalisis / dikerjakan
- `done` = sudah dirapikan dan/atau diberi alias kompatibilitas
- `legacy_exception` = sengaja dipertahankan seperti old contract
- `accepted_mixed` = kontrak campuran diterima dan didokumentasikan, tidak dipaksa diseragamkan saat ini

## Batch Kerja
### Batch A - Compatibility & naming
Fokus:
- singular vs plural
- underscore vs minus
- alias route kompatibilitas
- endpoint lama tetap hidup bila perlu

Target awal:
- `sale-products-combo` vs `sales-products-combo`
- `daily_asset` vs `daily-assets`
- `user-branches` dashed params alias

### Batch B - User/System contracts
Fokus:
- area user detail sederhana vs detail kaya
- branch / user-branch contract clarity
- auth legacy yang tetap dipertahankan

### Batch C - Master/supporting contracts
Fokus:
- combo/master routes yang masih ambigu
- perapihan naming bila benar-benar perlu dan aman

### Batch D - Transaction support surface
Fokus:
- combo/detail/export route pendukung transaksi
- tanpa menyentuh opname/opname-item

## Baseline Keputusan yang Sudah Ada
### done
- `GET /api/sale-products-combo` alias kompatibilitas aktif ke combo sale dan runtime tervalidasi
- alias dashed untuk `user-branches/:user-id/:branch-id`
- `GET /api/users/:user_id` sudah direpair sebagai detail sederhana dan runtime tervalidasi
- `GET /api/daily-assets` alias kompatibilitas list sudah aktif di atas `GET /api/daily_asset`

### legacy_exception
- `GET /api/opname-items` dengan body `{ "opname_id": "..." }`
- seluruh area `opname` / `opname-item`

### accepted_mixed
- export daily asset aktif di `/api/daily-assets/excel` sebagai kontrak plural yang tetap dipertahankan
- area `users` vs `detail-users` diterima sebagai kontrak ganda dengan level detail berbeda


### Batch D - Transaction support surface progress
- `GET /api/cmb-prod-sale-returns` empty-state dinormalisasi agar mengembalikan `200` + `data: []` saat tidak ada item retur yang tersedia.

- `GET /api/sale-returns` message list dikoreksi agar menyebut retur penjualan, bukan retur pembelian.

- Area `duplicate receipt` dibersihkan dari beberapa message copy yang masih menyebut `sale` / `sale item`.

- Area `expenses` dan `another-incomes` dirapikan agar response text lebih konsisten, domain-aware, dan lebih siap dikonsumsi frontend.

- Area `purchases` dirapikan agar message list/detail/item-level lebih konsisten dan domain-aware.

- Area `sales` dirapikan agar message list/detail/item-level lebih konsisten dan domain-aware.

- Area `duplicate-receipts` dirapikan lagi agar message list/detail/item-level lebih konsisten dan domain-aware.

- Area `duplicate_receipt` dilanjutkan ke round 3 untuk membersihkan response text user-facing pada jalur create/update/delete dan item-level.

- Area `sale` dilanjutkan ke lapis kedua untuk membersihkan response text validasi, transaksi inti, dan item-level yang masih campuran.

- Area `purchase` dilanjutkan ke lapis kedua untuk membersihkan response text validasi, transaksi inti, item-level, fixed price, dan product-units support.

- Area `first_stock` dilanjutkan ke lapis kedua untuk membersihkan response text validasi, transaksi inti, list/detail, dan item-level.
