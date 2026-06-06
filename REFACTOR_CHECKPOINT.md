# Refactor Checkpoint

Tanggal checkpoint: 2026-06-05
Fokus fase ini: stabilisasi runtime dan refactor bertahap modul transaksi utama tanpa mengubah kontrak API.

## Ringkasan eksekutif

Checkpoint ini menandai bahwa modul transaksi utama sudah masuk fase refactor bertahap yang relatif aman:
- runtime inti sudah tervalidasi
- export utama sudah aman
- purchase sudah menjadi template refactor paling matang saat ini
- sale sudah mulai mengikuti pola yang sama
- duplicate receipt sudah mulai mengikuti pola yang sama dan wiring helper dasarnya sudah selesai

Pendekatan yang dipakai selama fase ini:
1. pertahankan behavior API
2. pecah helper kecil dulu
3. build + smoke test setelah tiap langkah aman
4. commit + push langsung ke `main`

## Runtime baseline yang sudah aman

Sudah tervalidasi pada runtime:
- auth dua tahap
- protected endpoints utama
- master data inti
- purchase
- sale
- buy return
- sale return
- duplicate receipt
- first stock
- opname basic
- reporting dasar
- export Excel/PDF utama
- export item-level
- export audit/report/dashboard

Dokumen baseline runtime tersedia di:
- `RUNTIME_AUDIT.md`

## Perbaikan runtime penting yang sudah pernah selesai

- Konsistensi validasi token/JWT setelah `set_branch`
- Route export tidak lagi tertabrak route dinamis `/:id`
- Response `first stock detail payment` sudah benar
- Smoke test transaksi utama dan stok movement sudah tervalidasi

## Progress refactor per modul

## 1. Purchase

Purchase adalah modul yang paling jauh direfactor di fase ini.

### Tujuan
- mengurangi kepadatan `purchase_handler.go`
- memindahkan kalkulasi dan lookup kecil ke helper reusable
- mulai merapikan orchestration tanpa mengubah response/kontrak API

### Helper dan perapihan yang sudah dilakukan
Di `services/purchase_service.go` sudah ditambahkan / digunakan pola berikut:
- `EnsurePurchaseEditable(...)`
- `ErrDataExpiredToEdit`
- `SumPurchaseItems(...)`
- `ParsePurchaseDate(...)`
- `PreparePurchaseItemValues(...)`
- `LookupPurchaseItemDependencies(...)`
- `BuildPurchasedProductUpdates(...)`
- `BuildPurchaseTransactionResponse(...)`
- `LookupPurchaseSupplier(...)`
- `BuildPurchaseItemResponse(...)`
- `BuildPurchaseItemModel(...)`

Di `purchase_handler.go` juga sudah ditambahkan helper lokal untuk merapikan flow:
- helper untuk item effect sync
- helper preparation per-item transaction
- helper rollback response
- helper transaction report / quota orchestration kecil

### Hasil
- `CreatePurchaseTransaction` sudah jauh lebih pendek dibanding awal
- loop item sudah lebih terstruktur
- rollback handling lebih konsisten
- error mapping lebih rapi di beberapa titik
- smoke test create/delete purchase transaction tetap lulus setelah tiap refactor

### Catatan
Behavior create purchase transaction sengaja belum diubah, termasuk observasi response:
- `total_purchase: 0`
- `items: null`

Perilaku ini masih dipertahankan demi kompatibilitas sambil menunggu keputusan eksplisit jika ingin dibenahi.

## 2. Sale

Sale sudah masuk fase refactor bertahap, mengikuti pola dari purchase.

### Helper / perapihan yang sudah dilakukan
Di `services/sale_service.go`:
- `EnsureSaleEditable(...)`
- `ErrSaleDataExpiredToEdit`
- `PreparedSaleTotals`
- `AddSaleItemContribution(...)`
- `SaleItemStockUpdate`
- `BuildSaleItemStockUpdate(...)`
- `SaleProductLookup`
- `LookupSaleProduct(...)`
- `ValidateSaleStock(...)`
- `PreparedSaleItem`
- `PrepareSaleItem(...)`

Di `sale_handler.go`:
- rollback response helper dipusatkan
- kalkulasi total/profit sale dipisah ke helper
- lookup product + stock validation dipisah
- stock update dipisah
- item preparation mulai dipadatkan
- orchestration transaction report + daily profit + quota mulai dipisah ke helper lokal

### Hasil
- `CreateSaleTransaction` lebih bersih dan lebih konsisten polanya dengan purchase
- smoke test create/delete sale tetap lulus setelah tiap refactor

### Catatan penting
Pernah terjadi regression compile kecil saat refactor orchestration sale karena `errors` belum diimport. Sudah diperbaiki dan dipush.

## 3. Duplicate Receipt

Duplicate receipt sudah berhasil masuk ke pola refactor yang sama, dan sekarang kedalamannya sudah lebih baik dibanding checkpoint sebelumnya.

### Helper / perapihan yang sudah dilakukan
Di `services/duplicate_receipt_service.go`:
- `EnsureDuplicateReceiptEditable(...)`
- `ErrDuplicateReceiptDataExpiredToEdit`
- `PreparedDuplicateReceiptTotals`
- `AddDuplicateReceiptContribution(...)`
- `DuplicateReceiptProductLookup`
- `LookupDuplicateReceiptProduct(...)`
- `ValidateDuplicateReceiptStock(...)`
- `DuplicateReceiptStockUpdate`
- `BuildDuplicateReceiptStockUpdate(...)`
- `PreparedDuplicateReceiptItem`
- `PrepareDuplicateReceiptItem(...)`

Di `duplicate_receipt_handler.go`:
- helper-helper dasar sudah diwiring ke create flow
- editability path sudah dipindah ke helper reusable
- orchestration helper sudah ditambahkan untuk:
  - rollback response
  - transaction report
  - daily profit
  - quota handling

### Hasil
- create/update/delete duplicate receipt tetap lulus setelah wiring dan refactor orchestration
- duplicate receipt sekarang punya helper reusable untuk editability, totals, product lookup, stock validation, stock update, item preparation, dan orchestration
- duplicate receipt sudah mulai punya bentuk yang sejalan dengan purchase dan sale tanpa menyatukan domain bisnisnya

### Catatan penting
Sama seperti sale, sempat ada regression compile kecil karena `errors` belum diimport setelah refactor orchestration. Sudah diperbaiki dan dipush.


## 4. Buy Return

Buy return sekarang sudah memiliki fondasi refactor awal yang sehat.

### Helper / perapihan yang sudah dilakukan
Di `services/buy_return_service.go`:
- `ParseBuyReturnDate(...)`
- `SumBuyReturnSubTotal(...)`
- `BuildBuyReturnStockReduction(...)`
- `BuildBuyReturnResponse(...)`
- `ValidateBuyReturnQuantity(...)`
- `LookupBuyReturnPurchaseItem(...)`
- `LookupBuyReturnReturnedQty(...)`

Di `buy_return_handler.go`:
- parsing tanggal dipisah
- subtotal dipisah
- actual qty reduction dipisah
- lookup purchase item dan histori retur dipisah
- rollback response helper dipusatkan
- transaction report helper dipisah
- quota helper dipisah

### Hasil
- create flow tetap lolos validasi domain yang benar
- refactor buy return tidak mengubah rule qty retur
- buy return sekarang sudah jauh lebih sejajar dengan pola purchase/sale/duplicate receipt

## 5. Sale Return

Sale return juga sudah berhasil masuk ke pola refactor awal yang sama dengan buy return.

### Helper / perapihan yang sudah dilakukan
Di `services/sale_return_service.go`:
- `ParseSaleReturnDate(...)`
- `SumSaleReturnSubTotal(...)`
- `ValidateSaleReturnQuantity(...)`
- `LookupSaleReturnSaleItem(...)`
- `LookupSaleReturnReturnedQty(...)`
- `BuildSaleReturnResponse(...)`

Di `sale_return_handler.go`:
- parsing tanggal dipisah
- lookup sale item dan histori retur dipisah
- validasi qty retur dipisah
- subtotal dipisah
- response akhir dipisah
- rollback response helper dipusatkan
- transaction report helper dipisah
- quota helper dipisah

### Hasil
- create flow tetap lolos validasi domain yang benar
- rule qty retur tetap aman
- sale return sekarang sudah punya fondasi refactor yang kuat untuk fase berikutnya

## Pola kerja yang terbukti aman

Pola refactor yang terbukti paling aman selama fase ini:
1. identifikasi blok duplikasi / kalkulasi / lookup kecil
2. extract helper kecil dulu
3. build
4. smoke test runtime untuk flow yang terdampak
5. commit + push
6. baru lanjut ke blok berikutnya

Pola ini lebih efektif dibanding refactor besar sekali jalan, karena:
- lebih mudah melacak regression
- lebih mudah rollback mental/modeling
- lebih cocok dengan codebase yang dependency-nya masih campur antara handler, helpers, services, dan reports

## Regression yang sempat terjadi dan pelajarannya

### 1. Import cycle saat helper terlalu cepat ditarik ke service
Terjadi saat helper service mencoba memakai `helpers.GenerateID(...)` atau dependency yang memutar import graph.

Pelajaran:
- helper service sebaiknya fokus pada data transformation, lookup, kalkulasi, atau validation ringan
- generation ID atau orchestration yang berpotensi menarik `helpers` lebih aman ditahan di handler / helper lokal dulu

### 2. Missing import `errors`
Terjadi saat orchestration helper sale/duplicate receipt memakai `errors.New(...)` tetapi import belum ditambahkan.

Pelajaran:
- setelah refactor fungsi lokal yang memakai sentinel errors, selalu build langsung sebelum menganggap perubahan aman

## Yang masih belum selesai

### Purchase
- masih belum sepenuhnya kecil
- masih ada ruang memindahkan sebagian helper lokal ke service/usecase yang lebih matang bila dependency graph nanti memungkinkan

### Sale
- masih bisa dirapikan lagi, terutama area orchestration lanjutan dan item flow yang lebih dalam

### Duplicate receipt
- sudah cukup sehat dan jauh lebih rapi daripada titik awal

### Buy return
- sudah punya fondasi refactor kuat, tetapi masih belum sedalam purchase

### Sale return
- sudah punya fondasi refactor kuat, tetapi masih belum sedalam purchase

### Modul transaksi lain
Belum masuk refactor mendalam fase ini:
- first stock
- kemungkinan expense / another income untuk tahap service/usecase lebih lanjut jika diperlukan

## Rekomendasi tahap selanjutnya

Urutan yang disarankan setelah checkpoint ini:
1. tentukan modul transaksi berikutnya yang mau diperdalam
2. lanjutkan dengan pola yang sama, kecil tapi konsisten
3. pertimbangkan kapan mulai mengekstrak helper lokal menjadi service/usecase yang lebih resmi

### Rekomendasi prioritas
Pilihan paling masuk akal setelah checkpoint ini:
- mulai `first stock`
- atau memperdalam `buy return` / `sale return` satu tingkat lagi bila dibutuhkan

Jika ingin menyeimbangkan coverage transaksi, opsi paling masuk akal sekarang adalah:
- lanjut ke `first stock`

Jika ingin menuntaskan pola sebelum pindah lebih jauh, opsi paling masuk akal adalah:
- satu lapis pendalaman tambahan di `buy return` atau `sale return`

## Kesimpulan

Fase ini berhasil menghasilkan tiga hal penting:
1. runtime baseline yang kuat
2. pola refactor incremental yang terbukti aman
3. tiga modul transaksi utama yang mulai terkonsolidasi ke pola yang sama: purchase, sale, duplicate receipt

Purchase saat ini adalah template refactor paling matang.
Sale berada di jalur yang benar.
Duplicate receipt sudah berhasil masuk ke jalur yang sama.
