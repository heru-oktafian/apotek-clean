# API_CONTRACT_DIFF.md

## Cara baca
Setiap entri membandingkan kontrak di old repo / Postman lama dengan kontrak aktif di new repo.

Format:
- **Old Repo**
- **New Repo**
- **Perubahan**
- **Kompatibilitas**
- **Catatan**

---

## 1. User Detail Contracts
### Old Repo
- `GET /api/detail-users/:id`
- `GET /api/users/:user_id` secara historis membingungkan dan sempat tidak sehat

### New Repo
- `GET /api/detail-users/:id` -> detail user yang kaya + branches
- `GET /api/users/:user_id` -> detail sederhana user

### Perubahan
- kontrak dipisah lebih jelas berdasarkan level detail

### Kompatibilitas
- `detail-users` tetap dipertahankan
- `users/:user_id` direpair, bukan dihapus

### Catatan
- area ini sekarang resolved di runtime

---

## 2. Sale Product Combo Compatibility
### Old Repo
- `GET /api/sale-products-combo?search=...`

### New Repo
- kontrak utama aktif: `GET /api/sales-products-combo?search=...`
- alias kompatibilitas: `GET /api/sale-products-combo?search=...`

### Perubahan
- new repo menegaskan surface utama plural, tetapi tetap menerima kontrak lama

### Kompatibilitas
- kompatibel non-breaking

### Catatan
- diputuskan karena Duplicate Receipt harus reuse combo Sale

---

## 3. Daily Asset Contracts
### Old Repo
- list historis mengarah ke pola yang membingungkan antara singular/plural
- export plural dipakai di Postman

### New Repo
- list aktif: `GET /api/daily_asset`
- list kompatibel: `GET /api/daily-assets`
- export aktif: `GET /api/daily-assets/excel`

### Perubahan
- singular lama tetap hidup
- plural list ditambahkan sebagai compatibility path
- plural export tetap dipertahankan

### Kompatibilitas
- kompatibel non-breaking

### Catatan
- area ini tidak lagi jadi gap besar untuk frontend

---

## 4. User Branch Route Params
### Old Repo
- `:user_id/:branch_id`

### New Repo
- legacy tetap hidup: `:user_id/:branch_id`
- alias baru: `:user-id/:branch-id`

### Perubahan
- penambahan alias dashed params untuk surface yang lebih rapi

### Kompatibilitas
- kompatibel non-breaking

### Catatan
- route handler membaca kedua bentuk param

---

## 5. Opname / Opname Item Exception
### Old Repo
- `GET /api/opname-items` dengan body `{ "opname_id": "..." }`

### New Repo
- kontrak lama tetap dipertahankan

### Perubahan
- tidak dinormalisasi

### Kompatibilitas
- 100% legacy-compatible

### Catatan
- area ini dikecualikan dari program normalisasi atas keputusan user


## 6. Runtime-verified compatibility status
### Verified active after runtime sync
- `GET /api/sale-products-combo?search=...` -> 200
- `GET /api/sales-products-combo?search=...` -> 200
- `GET /api/daily_asset?...` -> 200
- `GET /api/daily-assets?...` -> 200
- `GET /api/users/:user_id` -> 200
- `PUT /api/users/:user_id` -> 200
- `GET /api/detail-users/:id` -> 200

### Operational note
- Beberapa verifikasi sempat menipu saat proses lama yang listen di `:9002` masih memakai binary `(deleted)`.
- Status compatibility hanya dianggap final setelah PID listener benar-benar diganti ke binary terbaru.

---

## 6. User Branch Contract Clarity Note
### Old Repo
- route memakai bentuk `:user_id/:branch_id`
- ekspektasi alami klien: kedua path param ikut menentukan target relasi user-branch

### New Repo
- legacy route tetap hidup: `:user_id/:branch_id`
- alias dashed juga hidup: `:user-id/:branch-id`
- namun pada behavior aktif tertentu, branch context masih bergantung pada token aktif, bukan murni pada `branch_id` dari path

### Perubahan
- surface kompatibel sudah ditambahkan
- tetapi semantics path vs token context belum sepenuhnya intuitif

### Kompatibilitas
- kompatibel secara route
- masih ada clarity issue pada perilaku aktual

### Catatan
- belum diubah sekarang agar tidak memicu perubahan behavior tanpa keputusan eksplisit
- perlu keputusan terpisah bila ingin path `branch_id` benar-benar menjadi penentu utama target operasi


## 7. Sale Return Combo Empty-state Normalization
### Old Repo
- `GET /api/cmb-prod-sale-returns?sale_id=...` dapat mengembalikan `404` saat tidak ada item yang bisa diretur

### New Repo
- `GET /api/cmb-prod-sale-returns?sale_id=...` mengembalikan `200` dengan `data: []` saat kosong

### Perubahan
- empty-state combo dibuat lebih frontend-friendly dan konsisten sebagai response list

### Kompatibilitas
- non-breaking untuk consumer yang memproses sukses-list
- mengurangi kebutuhan penanganan khusus 404 untuk kondisi data kosong

### Catatan
- runtime tervalidasi setelah listener stale `:9002` dibersihkan dan binary terbaru dimuat


## 8. Sale Return List Message Correction
### Old Repo
- `GET /api/sale-returns` mengembalikan message `Data retur pembelian berhasil diambil`

### New Repo
- `GET /api/sale-returns` mengembalikan message `Data retur penjualan berhasil diambil`

### Perubahan
- koreksi message agar sesuai domain endpoint

### Kompatibilitas
- non-breaking, hanya memperbaiki kejelasan response text

### Catatan
- runtime tervalidasi setelah restart listener baru pada PID aktif yang bersih


## 9. Duplicate Receipt Message Cleanup
### Old Repo
- beberapa response/error text di area duplicate receipt masih memakai istilah `sale` / `sale item`

### New Repo
- response/error text dibersihkan agar menyebut `duplicate receipt` / `duplicate receipt item` sesuai domain

### Perubahan
- koreksi message internal-facing dan client-facing tanpa mengubah alur bisnis

### Kompatibilitas
- non-breaking, memperjelas domain response

### Catatan
- create/delete item diverifikasi runtime setelah listener kembali stabil


## 10. Expense and Another Income Message Cleanup
### Old Repo
- area `expenses` dan `another-incomes` masih memakai banyak response text campuran bahasa Inggris / kapitalisasi domain yang kurang rapi

### New Repo
- response text utama dibersihkan menjadi lebih konsisten dan domain-aware, misalnya `Data pengeluaran berhasil diambil` dan `Data pendapatan lain berhasil diambil`

### Perubahan
- perapihan message create/update/delete/list tanpa mengubah alur bisnis

### Kompatibilitas
- non-breaking, lebih ramah untuk frontend yang menampilkan message langsung

### Catatan
- runtime tervalidasi setelah listener stale dibersihkan dan PID baru memuat binary terbaru


## 11. Purchase Response Message Cleanup
### Old Repo
- area `purchases` masih memakai beberapa response text berbahasa Inggris seperti `Purchases retrieved successfully`, `Items retrieved successfully`, dan `Purchase retrieved successfully`

### New Repo
- response text utama dibersihkan menjadi lebih konsisten, misalnya `Data pembelian berhasil diambil` dan `Data item pembelian berhasil diambil`

### Perubahan
- perapihan message list/detail/item-level/create-update-delete tanpa mengubah alur bisnis

### Kompatibilitas
- non-breaking, lebih enak untuk frontend yang menampilkan message secara langsung

### Catatan
- verifikasi akhir sempat tertunda karena hasil test awal belum sinkron; binary terbaru kemudian dipastikan memuat string baru dan runtime final tervalidasi 200


## 12. Sale Response Message Cleanup
### Old Repo
- area `sales` masih memakai beberapa response text berbahasa Inggris seperti `Sales retrieved successfully`, `Items retrieved successfully`, dan `Sale retrieved successfully`

### New Repo
- response text utama dibersihkan menjadi lebih konsisten, misalnya `Data penjualan berhasil diambil` dan `Data item penjualan berhasil diambil`

### Perubahan
- perapihan message list/detail/item-level/create-update-delete tanpa mengubah alur bisnis

### Kompatibilitas
- non-breaking, lebih enak untuk frontend yang menampilkan message secara langsung

### Catatan
- runtime tervalidasi untuk `sales`, `sales-details`, `sale-items/all/:id`, dan `sales/:id` setelah restart listener bersih


## 13. Duplicate Receipt Response Message Cleanup Round 2
### Old Repo
- area `duplicate-receipts` masih memakai beberapa response text berbahasa Inggris seperti `Duplicate receipts retrieved successfully`, `Items retrieved successfully`, dan `Duplicate receipt retrieved successfully`

### New Repo
- response text utama dibersihkan menjadi lebih konsisten, misalnya `Data duplicate receipt berhasil diambil` dan `Data item duplicate receipt berhasil diambil`

### Perubahan
- perapihan message list/detail/item-level tanpa mengubah alur bisnis

### Kompatibilitas
- non-breaking, lebih enak untuk frontend yang menampilkan message secara langsung

### Catatan
- runtime tervalidasi untuk `duplicate-receipts`, `duplicate-receipts-details`, `duplicate-receipts-items/all/:id`, dan `duplicate-receipts/:id`


## 14. Duplicate Receipt Error and Item Message Cleanup Round 3
### Old Repo
- area `duplicate-receipt` masih memiliki beberapa response text campuran seperti `Invalid request body`, `Invalid input`, `Duplicate receipt updated successfully`, `Item updated successfully`, dan `Item deleted successfully`

### New Repo
- response text create/update/delete/item-level dibersihkan menjadi lebih konsisten dan domain-aware

### Perubahan
- koreksi message user-facing pada jalur body parsing, update duplicate receipt, create/update/delete item duplicate receipt

### Kompatibilitas
- non-breaking, tetap mempertahankan flow bisnis yang sama

### Catatan
- detail duplicate receipt serta flow create/delete item tervalidasi runtime setelah restart listener bersih


## 15. Sale Error and Support Message Cleanup Round 2
### Old Repo
- area `sale` masih menyisakan beberapa response text teknis berbahasa Inggris pada jalur validasi, transaksi, dan item-level seperti `Validate failed`, `Failed to begin database transaction`, `Failed to create sale item`, dan sejenisnya

### New Repo
- response text user-facing pada jalur validasi, transaksi inti, sinkronisasi report, dan item-level dirapikan ke Bahasa Indonesia yang lebih konsisten

### Perubahan
- koreksi message validasi input penjualan
- koreksi message transaksi create/update sale
- koreksi message pendukung item-level dan report sync

### Kompatibilitas
- non-breaking, hanya merapikan kontrak message tanpa mengubah flow bisnis

### Catatan
- list sale dan detail sale tervalidasi runtime pada listener bersih setelah restart


## 16. Purchase Error and Support Message Cleanup Round 2
### Old Repo
- area `purchase` masih menyisakan response text teknis berbahasa Inggris pada jalur create/update/delete, item-level, purchase transaction, fixed price, dan product-units support

### New Repo
- response text user-facing dirapikan ke Bahasa Indonesia yang lebih konsisten pada jalur-jalur support tersebut

### Perubahan
- koreksi message validasi input pembelian dan body parsing
- koreksi message create/update purchase dan item-level
- koreksi message purchase transaction, quota, fixed price, dan product-units support

### Kompatibilitas
- non-breaking, hanya merapikan kontrak message tanpa mengubah flow bisnis

### Catatan
- list purchase, detail purchase, dan item list purchase tervalidasi runtime pada listener bersih setelah restart


## 17. First Stock Error and Support Message Cleanup Round 2
### Old Repo
- area `first_stock` masih menyisakan response text teknis berbahasa Inggris pada jalur create/update/delete, item-level, list/detail, dan transaksi first stock

### New Repo
- response text user-facing dirapikan ke Bahasa Indonesia yang lebih konsisten pada jalur-jalur support tersebut

### Perubahan
- koreksi message validasi input, format tanggal, dan body parsing
- koreksi message create/update/delete first stock dan item-level
- koreksi message list/detail first stock dan transaksi first stock

### Kompatibilitas
- non-breaking, hanya merapikan kontrak message tanpa mengubah flow bisnis

### Catatan
- list, detail, item list, dan create/delete item first stock tervalidasi runtime pada listener bersih setelah restart


## 18. Unit Conversion Error and Support Message Cleanup Round 2
### Old Repo
- area `unit_conversion` masih menyisakan response text teknis berbahasa Inggris pada jalur create, validasi dependency, pagination list, dan combo product conversion

### New Repo
- response text user-facing dirapikan ke Bahasa Indonesia yang lebih konsisten pada jalur-jalur support tersebut

### Perubahan
- koreksi message branch/token, begin/commit transaksi, create unit conversion, conflict duplicate, dan lookup dependency
- koreksi message list unit conversion dan combo product conversion

### Kompatibilitas
- non-breaking, hanya merapikan kontrak message tanpa mengubah flow bisnis

### Catatan
- list unit conversion tervalidasi runtime `200` dalam kondisi empty-state (`data: null`) dan combo product conversion tervalidasi runtime `200`
