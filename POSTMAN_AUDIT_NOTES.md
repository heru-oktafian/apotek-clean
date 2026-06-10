# Postman Audit Notes

## 1. Match / aman dipertahankan
- `POST /api/login`
- `GET /api/list_branches`
- `POST /api/set_branch`
- `GET /api/profile`
- `GET /api/branches`
- `GET /api/detail-users/:id`
- `GET /api/sys-defectas`
- `GET /api/products`
- `GET /api/purchases`
- `GET /api/sales-details`
- `GET /api/duplicate-receipts-details`
- `GET /api/expenses`
- `GET /api/another-incomes`
- `GET /api/first-stock-items/:id`
- `GET /api/opnames`
- `GET /api/daily_asset`
- `GET /api/report/neraca-saldo`
- `GET /api/report/profit-by-month`
- `GET /api/dashboard/top-selling-report`
- `GET /api/dashboard/least-selling-report`
- `GET /api/dashboard/neared-report`

## 2. Kontrak aktif yang perlu didokumentasikan jelas
- Detail user aktif adalah `GET /api/detail-users/:id`, bukan `GET /api/users/:id`.
- Route daily asset aktif untuk list adalah `GET /api/daily_asset`, sedangkan export tetap `GET /api/daily-assets/excel`.
- Sales list aktif yang tervalidasi adalah `GET /api/sales-details`, bukan asumsi `GET /api/sales` untuk list detail tabel.
- Duplicate receipt list detail aktif yang tervalidasi adalah `GET /api/duplicate-receipts-details`.

## 3. Kandidat normalisasi / perhatian lanjutan
- `GET /api/opname-items`
  - Route aktif ada, tetapi handler membaca `opname_id` dari body JSON.
  - Ini kontrak yang ambigu dan tidak lazim untuk GET.
  - Kandidat normalisasi: pindah ke query param `?opname_id=` atau route path yang eksplisit.
- `GET /api/users/:id`
  - Bukan kontrak detail user yang benar.
  - Jika tetap dibiarkan aktif, berisiko membingungkan klien lama/baru.
  - Kandidat normalisasi: dokumentasikan deprecated atau redirect kontrak internal di lapisan route.

## 4. Rekomendasi tahap awal
1. Dokumentasikan kontrak aktif yang sudah pasti benar.
2. Tandai endpoint ambigu sebagai kandidat normalisasi, jangan diubah dulu tanpa keputusan eksplisit.
3. Prioritas normalisasi pertama yang paling masuk akal: `opname-items`.

## 5. Keputusan khusus: pengecualian opname
- Untuk area `opname` dan `opname-item`, kontrak legacy dipertahankan.
- Route seperti `GET /api/opname-items` tetap dibiarkan seperti perilaku lama dan tidak dinormalisasi.
- Upaya penambahan route baru untuk normalisasi sudah dibatalkan agar tidak mengubah kontrak yang diinginkan user.

## 6. Temuan lanjutan dari verifikasi runtime
- `GET /api/sales` -> 200
- `GET /api/sales-details` -> 200
- `GET /api/duplicate-receipts` -> 200
- `GET /api/duplicate-receipts-details` -> 200
- `GET /api/daily_asset` -> 200
- `GET /api/daily-assets` -> 404
- `GET /api/detail-users/:id` -> 200
- `GET /api/users/:id` -> 404 / tidak cocok untuk kontrak detail user

## 7. Normalisasi underscore ke minus: hasil klasifikasi awal
### Biarkan dulu
- `GET /api/list_branches`
- `POST /api/set_branch`

### Jangan disentuh dulu
- `GET /api/daily_asset`
- `GET|PUT|DELETE /api/users/:user_id`

### Kandidat aman untuk alias minus
- `GET|PUT|DELETE /api/user-branches/:user_id/:branch_id`

## 8. Eksekusi normalisasi selektif yang sudah dilakukan
- Alias dashed sudah ditambahkan untuk area user-branches tanpa mematikan kontrak lama:
  - `GET /api/user-branches/:user-id/:branch-id`
  - `PUT /api/user-branches/:user-id/:branch-id`
  - `DELETE /api/user-branches/:user-id/:branch-id`
- Kontrak lama tetap dipertahankan:
  - `GET /api/user-branches/:user_id/:branch_id`
  - `PUT /api/user-branches/:user_id/:branch_id`
  - `DELETE /api/user-branches/:user_id/:branch_id`

## 9. Verifikasi final setelah restart runtime yang benar
- Ditemukan bahwa beberapa pengecekan sebelumnya sempat menipu karena proses lama yang listen di `:9002` memakai binary `(deleted)` lama.
- Setelah proses lama dimatikan dan binary terbaru dijalankan ulang, hasil final menjadi:
  - `GET /api/sale-products-combo?search=gliben` -> 200
  - `GET /api/sales-products-combo?search=gliben` -> 200
  - `GET /api/users/USR250118132201` -> 200
  - `GET /api/detail-users/USR250118132201` -> 200
- Artinya:
  - alias kompatibilitas `sale-products-combo` sudah aktif
  - repair `GET /api/users/:user_id` sudah aktif
  - `detail-users` tetap sehat sebagai kontrak detail user yang lebih kaya

## 10. Catatan operasional penting
- Bila hasil runtime tampak bertentangan dengan source terbaru, cek dulu proses yang listen di `:9002`.
- Kasus nyata yang terjadi: proses lama masih memakai executable `/home/jarvis/.dev/apotek-clean/bin/apotek (deleted)`.
- Pola aman verifikasi runtime:
  1. matikan PID yang benar-benar listen di `:9002`
  2. build ulang binary `./bin/apotek`
  3. start ulang
  4. pastikan PID baru yang listen memang binary terbaru

## 11. Klarifikasi false alarm pada PDF detail item
- Tiga endpoint berikut sempat terlihat `500` saat diuji dengan sample ID lama/tidak valid:
  - `GET /api/first-stock-items/pdf?first_stock_id=...`
  - `GET /api/buy-return-items/pdf?buy_return_id=...`
  - `GET /api/sale-return-items/pdf?sale_return_id=...`
- Setelah diuji ulang menggunakan ID valid dari branch aktif, hasil final menjadi:
  - `first-stock-items/pdf` -> 200 / `application/pdf`
  - `buy-return-items/pdf` -> 200 / `application/pdf`
  - `sale-return-items/pdf` -> 200 / `application/pdf`
- Kesimpulan: ini bukan bug code atau mismatch route, melainkan false alarm akibat sample data lama yang tidak cocok dengan branch/token aktif.

## 12. Aturan audit runtime yang dipertegas
- Untuk endpoint detail/export yang bergantung pada parent transaction ID, jangan simpulkan bug hanya dari sample Postman lama.
- Selalu ambil ID valid lebih dulu dari branch aktif sebelum menyimpulkan 500 sebagai bug runtime.

## 13. Final gap summary
### Real mismatch / perlu diingat
- List Daily Asset aktif: `GET /api/daily_asset`
- `GET /api/daily-assets` tidak aktif untuk list (plural tetap dipakai untuk export tertentu)
- `GET /api/detail-users/:id` adalah kontrak detail user yang kaya (user + branches)
- `GET /api/users/:user_id` kini hidup sebagai detail sederhana, bukan pengganti `detail-users`

### Legacy contract yang sengaja dipertahankan
- `GET /api/opname-items` dengan body `{"opname_id":"..."}`
- Seluruh pola `opname` / `opname-item` tidak dinormalisasi route sesuai keputusan user
- Flow auth legacy `GET /api/list_branches` dan `POST /api/set_branch` tetap dipertahankan

### Kompatibilitas / alias yang sudah ditambahkan
- `GET /api/sale-products-combo` -> alias kompatibilitas ke combo sale aktif
- `GET|PUT|DELETE /api/user-branches/:user-id/:branch-id` -> alias dashed di atas route lama `:user_id/:branch_id`

### False alarm yang sudah terpatahkan
- `first-stock-items/pdf` sempat 500 karena sample ID lama tidak valid
- `buy-return-items/pdf` sempat 500 karena sample ID lama tidak valid
- `sale-return-items/pdf` sempat 500 karena sample ID lama tidak valid
- Verifikasi runtime sempat menipu karena proses lama di `:9002` memakai binary `(deleted)`

## 14. Status final daily asset
- Kontrak aktif yang diterima saat ini:
  - list: `GET /api/daily_asset`
  - export: `GET /api/daily-assets/excel`
- `GET /api/daily-assets` untuk list tetap tidak aktif.
- Ini diperlakukan sebagai kontrak campuran yang diterima, bukan bug runtime kritis.
- Keputusan: didokumentasikan apa adanya, tidak dipaksa seragam untuk saat ini.

## 15. Status final area user setelah runtime benar-benar sinkron
- Setelah proses lama yang masih memegang `:9002` dimatikan dan binary baru benar-benar jalan, verifikasi final area user menjadi:
  - `GET /api/users/USR250118132201` -> 200
  - `PUT /api/users/USR250118132201` -> 200
  - `GET /api/detail-users/USR250118132201` -> 200
- Kesimpulan kontrak area user:
  - `GET /api/users/:user_id` = detail sederhana user
  - `GET /api/detail-users/:id` = detail user yang lebih kaya + branches
- Error lama `column "user_id" does not exist` pada update user dinyatakan resolved setelah runtime benar-benar memuat source terbaru.


- `GET /api/cmb-prod-sale-returns?sale_id=...` dinormalisasi dari empty-state `404` menjadi `200` dengan `data: []` agar konsisten sebagai combo/list endpoint. Verifikasi final dilakukan setelah proses stale `(deleted)` di `:9002` dibunuh dan listener baru dimuat.

- `GET /api/sale-returns` semula masih memakai message copy yang salah (`retur pembelian`). Sudah dikoreksi menjadi `Data retur penjualan berhasil diambil` dan tervalidasi runtime.

- Area `duplicate receipt` dirapikan dari response text yang masih copy dari domain sale, misalnya `Failed to delete sale` menjadi `Failed to delete duplicate receipt` dan sejenisnya. Flow create/delete item tervalidasi runtime setelah listener stabil.

- `GET /api/expenses` dan `GET /api/another-incomes` sudah menampilkan message baru yang lebih konsisten (`Data pengeluaran berhasil diambil`, `Data pendapatan lain berhasil diambil`) setelah restart listener bersih.

- `GET /api/purchases`, `GET /api/purchase-items/all/:id`, dan `GET /api/purchases/:id` sudah menampilkan message baru yang lebih konsisten (`Data pembelian berhasil diambil`, `Data item pembelian berhasil diambil`) setelah verifikasi ulang pada runtime final.

- `GET /api/sales`, `GET /api/sales-details`, `GET /api/sale-items/all/:id`, dan `GET /api/sales/:id` sudah menampilkan message baru yang lebih konsisten (`Data penjualan berhasil diambil`, `Data item penjualan berhasil diambil`) setelah restart listener bersih.

- `GET /api/duplicate-receipts`, `GET /api/duplicate-receipts-details`, `GET /api/duplicate-receipts-items/all/:id`, dan `GET /api/duplicate-receipts/:id` sudah menampilkan message baru yang lebih konsisten (`Data duplicate receipt berhasil diambil`, `Data item duplicate receipt berhasil diambil`) setelah restart listener bersih.

- Area `duplicate_receipt` round 3 dirapikan lagi pada jalur body parsing, update duplicate receipt, serta create/update/delete item. Flow detail dan create/delete item tervalidasi runtime setelah restart listener bersih.

- Area `sale` lapis kedua dirapikan pada jalur validasi input, transaksi inti, sinkronisasi report, dan item-level. List sale dan detail sale tervalidasi runtime setelah restart listener bersih.

- Area `purchase` lapis kedua dirapikan pada jalur create/update/delete, item-level, purchase transaction, fixed price, dan product-units support. List, detail, dan item list purchase tervalidasi runtime setelah restart listener bersih.

- Area `first_stock` lapis kedua dirapikan pada jalur create/update/delete, list/detail, item-level, dan transaksi first stock. List, detail, item list, serta create/delete item tervalidasi runtime setelah restart listener bersih.

- Area `unit_conversion` lapis kedua dirapikan pada jalur create, dependency validation, list pagination, dan combo product conversion. List empty-state dan combo tervalidasi runtime `200`.
