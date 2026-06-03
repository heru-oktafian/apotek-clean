# Runtime Audit

Tanggal audit: 2026-06-03
Environment uji: local app di `http://127.0.0.1:9002`
Akun uji: `vita_fauzi`
Branch uji: `BRC250118132203` / `Ziida Farma`

## Ringkasan

Audit ini memverifikasi bahwa hasil refactor route/handler internal tetap mempertahankan perilaku API utama pada runtime. Fokus audit mencakup auth dua tahap, master data, transaksi inti, inventory/stock movement, reporting, dan export Excel/PDF.

## Status Umum

- Build aplikasi: lulus
- Startup aplikasi: lulus
- Auth 2 tahap: lulus
- Endpoint protected tanpa token: mengembalikan `401`
- Route internal hasil migrasi: aktif
- Export routes: aktif setelah perbaikan urutan registrasi route

## Auth

Flow yang tervalidasi:
1. `POST /api/login` menghasilkan token awal
2. `GET /api/list_branches` berhasil dengan token awal
3. `POST /api/set_branch` menghasilkan token branch-scoped
4. Endpoint protected utama berhasil diakses dengan token hasil `set_branch`

Catatan:
- JWT generator/validator/claim extraction sudah konsisten memakai fallback `JWT_SECRET` -> `JWT_SECRET_KEY`
- Bug token invalid setelah `set_branch` sudah diperbaiki sebelumnya

## Endpoint dasar protected

Tervalidasi mengembalikan `200` setelah auth branch-scoped:
- `/api/profile`
- `/api/branches`
- `/api/users`
- `/api/products`
- `/api/purchases`
- `/api/sales`
- `/api/dashboard/daily-profit-report`
- `/api/report/neraca-saldo`

## Master data

CRUD ringan tervalidasi sukses untuk:
- product categories
- supplier categories
- member categories
- suppliers
- members

## Transaksi ringan

Tervalidasi sukses:
- `expenses` create/update/delete
- `another-incomes` create/update/delete

Perbaikan yang sudah dilakukan:
- update expense tidak lagi mengirim `payment=""` ke enum DB
- update another income tidak lagi menimpa field sensitif saat kosong
- route `POST /api/another-incomes` tetap kompatibel

## Inventory and stock movement

Produk uji utama:
- `PRD028055HKY6YL` (`BOTOL DOT PIGEON 120 ML`)

Pergerakan stok yang tervalidasi:
- purchase: stok bertambah
- buy return: stok berkurang
- sale: stok berkurang
- sale return: stok bertambah
- first stock: stok bertambah
- duplicate receipt: stok berkurang
- delete sale: stok rollback naik kembali
- delete duplicate receipt: stok rollback naik kembali

Kesimpulan:
- arah perubahan stok antar transaksi inti konsisten dengan aturan bisnis

## Transaksi inti

### Purchase
- create: berhasil
- detail: berhasil
- delete transaksi baru: berhasil
- update transaksi lama: ditolak `403`

Catatan perilaku:
- Penolakan update purchase lama disebabkan business rule: data tidak bisa diedit setelah lebih dari 1 jam
- Ini dianggap expected behavior, bukan bug
- Pada skenario create temp tertentu, response create purchase menampilkan header dengan `total_purchase: 0` dan `items: null`; perilaku ini perlu diingat saat membuat smoke test, karena kontrak response berbeda dengan sale/duplicate receipt

### Sale
- create: berhasil
- detail: berhasil
- update: berhasil
- delete transaksi baru: berhasil
- rollback stok saat delete: berhasil

### Buy return
- create: berhasil
- list/detail: berhasil
- validasi qty return melebihi pembelian: berhasil menolak

### Sale return
- create: berhasil
- list/detail: berhasil
- kontrak payload tervalidasi memakai `sale_items`
- enum payment valid misalnya `paid_by_cash`

### Duplicate receipt
- create: berhasil
- detail: berhasil
- list: berhasil
- detail list: berhasil
- update: berhasil
- delete transaksi baru: berhasil
- rollback stok saat delete: berhasil
- sinkron ke daily profit report: berhasil

## Audit modules

### First stock
- create: berhasil
- add item: berhasil
- detail with items: berhasil
- stok bertambah sesuai item

Bug yang ditemukan dan diperbaiki:
- field `payment` pada detail first stock sempat salah berisi tanggal
- akar masalah selesai setelah patch response + memastikan binary terbaru yang berjalan

### Opname
- list: berhasil
- mobile list: berhasil
- combobox product opname: berhasil
- stok terbaru tercermin pada combobox

## Reporting

Tervalidasi berhasil:
- `/api/dashboard/daily-profit-report`
- `/api/dashboard/profit-today-by-user`
- `/api/report/neraca-saldo`

Catatan kecil:
- Pada `profit-today-by-user`, pernah teramati `user_name` terisi tetapi `user_id` kosong. Ini bukan blocker utama, tetapi layak dicatat sebagai anomali mapping report yang bisa dirapikan nanti.

## Export

### Akar masalah yang ditemukan
Beberapa endpoint export awalnya mengembalikan JSON biasa, bukan file, karena route seperti `/api/products/:id` menangkap `/api/products/excel` dan `/api/products/pdf` lebih dulu.

Perbaikan yang dilakukan:
- urutan registrasi route diubah agar export routes didaftarkan sebelum route dinamis `/:id`

### Export utama tervalidasi
Excel/PDF berikut menghasilkan file valid dengan content type dan magic bytes yang sesuai:
- products
- purchases
- sales
- duplicate receipts

### Export item-level tervalidasi
- purchase items (excel/pdf)
- sale items (excel/pdf)
- duplicate receipt items (excel/pdf)
- first stock items (excel/pdf)

### Export audit/report/dashboard tervalidasi
- first stocks (excel/pdf)
- opnames (excel/pdf)
- neraca saldo (excel)
- top selling report (excel)
- least selling report (excel)
- neared report (excel)
- daily assets (excel)
- defectas (excel)

## Known behaviors / catatan penting

1. Auth memang 2 tahap. Token login awal belum cukup untuk akses semua endpoint branch-scoped.
2. `.env` masih dibaca relatif terhadap working directory saat proses start.
3. False alarm bisa terjadi bila binary lama masih listen di `:9002`.
4. Update purchase memiliki time-lock > 1 jam.
5. Beberapa smoke test transaksi memerlukan data induk valid dan urutan setup yang tepat.

## Kesimpulan

Refactor route/handler internal saat ini sudah tervalidasi cukup kuat pada runtime untuk area penting:
- auth
- master data inti
- transaksi inti
- inventory/stok
- reporting dasar
- export utama dan detail

Masih ada ruang refactor lebih dalam ke service/usecase/domain, tetapi baseline runtime sudah jauh lebih aman sebagai pijakan tahap berikutnya.
