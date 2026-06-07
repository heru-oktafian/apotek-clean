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

Buy return sekarang sudah mencapai fase yang jauh lebih matang dan kualitasnya sudah sejajar dengan modul transaksi utama lain.

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
- buy return sekarang sudah memiliki helper dasar, loop item yang sudah dibersihkan, rollback yang terpusat, dan orchestration yang rapi
- buy return sudah layak dianggap hijau penuh untuk fase ini

## 5. Sale Return

Sale return juga sudah berhasil mencapai kualitas yang sejajar dengan buy return.

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
- sale return sekarang sudah memiliki helper dasar, loop item yang lebih rapi, rollback terpusat, dan orchestration yang konsisten
- sale return sudah layak dianggap hijau penuh untuk fase ini


## 6. First Stock

First stock sekarang sudah berhasil didorong sampai kualitasnya sejajar dengan modul transaksi utama lain.

### Helper / perapihan yang sudah dilakukan
Di `services/first_stock_service.go`:
- `EnsureFirstStockEditable(...)`
- `ErrFirstStockDataExpiredToEdit`
- `ParseFirstStockDate(...)`
- `SumFirstStockItems(...)`
- `LookupFirstStockDependencies(...)`
- `PrepareFirstStockItem(...)`

Di `first_stock_handler.go`:
- parsing tanggal dipisah
- sum total dipisah
- lookup dependency item dipisah
- item preparation dipisah
- rollback response helper dipusatkan
- transaction finalization helper dipisah
- quota/branch finalization dipisah

### Hasil
- create/delete first stock tetap lulus
- response detail dan item tetap sehat
- first stock sekarang sudah memiliki helper dasar, loop item yang bersih, rollback terpusat, dan orchestration yang rapi
- first stock sudah layak dianggap hijau penuh untuk fase ini


## 7. Opname

Opname sekarang sudah mulai keluar dari kondisi handler besar yang mentah, dan sudah punya fondasi refactor non-transaksional yang cukup sehat.

### Helper / perapihan yang sudah dilakukan
Di `services/opname_service.go`:
- `ParseOpnameDate(...)`
- `ParseOpnameItemDate(...)`
- `BuildOpnameItemSnapshot(...)`
- `PrepareOpnameItem(...)`
- `PrepareOpnameItemUpdate(...)`

Di `opname_handler.go`:
- formatting mobile opname dipindah ke helper lokal
- parsing tanggal create/update dipisah
- item preparation create dipisah
- orchestration create/update dipusatkan
- item update flow dipisah ke helper preparation

### Hasil
- `GET /api/mobile-opnames` tervalidasi
- `POST /api/opnames` tervalidasi
- `POST /api/opname-items` tervalidasi
- `PUT /api/opname-items` tervalidasi
- `PUT /api/opnames/:id` tervalidasi
- opname sekarang jauh lebih sehat dibanding kondisi awal dan cukup layak sebagai baseline non-transaksional yang sudah ditata


## 8. Daily Asset

Daily asset sekarang sudah mulai disentuh dengan refactor kecil yang aman, dan baseline runtime-nya juga sudah terkonfirmasi.

### Helper / perapihan yang sudah dilakukan
Di `services/daily_asset_service.go`:
- `ParseDailyAssetMonth(...)`
- `CalculateDailyAssetTotalPages(...)`

Di `daily_asset_handler.go`:
- parsing bulan dipisah
- perhitungan total pages dipisah
- formatting tanggal asset tetap ditahan di handler agar tidak memicu import cycle

### Hasil
- build sudah bersih setelah repair import cycle
- route aktif terkonfirmasi menggunakan `/api/daily_asset`
- `GET /api/daily_asset?month=YYYY-MM&page=1` tervalidasi runtime
- daily asset sekarang punya baseline refactor yang aman untuk dilanjutkan nanti bila dibutuhkan


## 9. Defecta

Defecta sudah mulai disentuh dengan pola refactor kecil yang aman, dan cleanup pentingnya sudah berhasil diselesaikan.

### Helper / perapihan yang sudah dilakukan
Di `services/defecta_service.go`:
- `ParseDefectaDate(...)`
- `SumDefectaSubTotal(...)`
- `BuildDefectaItemsResponse(...)`
- `CalculateDefectaTotalPages(...)`

Di `defecta_handler.go`:
- parsing tanggal mulai dipisah
- subtotal mulai dipisah
- formatting response list/detail dibantu helper service yang aman
- import cycle yang sempat muncul sudah diperbaiki

### Hasil
- `POST /api/sys-defectas` tervalidasi runtime
- `GET /api/sys-defectas` tervalidasi runtime
- build sudah bersih setelah repair final
- defecta sekarang cukup sehat untuk fase saat ini

## 10. Report

Report sudah mulai disentuh secara kecil dan pragmatis, terutama pada area date-range parsing.

### Helper / perapihan yang sudah dilakukan
Di `services/report_service.go`:
- `ParseYearMonthRange(...)`
- `ParseReportMonthBounds(...)`

Di `report_handler.go`:
- parsing range bulan untuk laporan bulanan dipindah ke helper reusable

### Hasil
- `GET /api/report/neraca-saldo?month=YYYY-MM` tervalidasi runtime
- `GET /api/report/profit-by-month?month=YYYY-MM` tervalidasi runtime
- report saat ini dianggap cukup sehat untuk fase sekarang dan tidak perlu dibongkar lebih jauh dulu


## 11. Expense

Expense sekarang sudah mulai punya fondasi refactor kecil yang aman dan cukup sehat untuk modul cashflow ringan.

### Update consistency pass
- Setelah audit internal consistency, `expense` dipilih sebagai kandidat ROI tertinggi untuk diratakan dengan `another_income`.
- `services/expense_service.go` sekarang mencakup:
  - `ParseExpenseDate(...)`
  - `NormalizeExpensePayment(...)`
  - `EnsureExpenseEditable(...)`
  - `ErrExpenseDataExpiredToEdit`
- `expense_handler.go` sekarang memakai helper tersebut untuk create/update flow.
- Runtime verification terbaru:
  - `POST /api/expenses` -> 200
  - `PUT /api/expenses/:id` -> 200
  - `DELETE /api/expenses/:id` -> 200
- Commit terkait: `refactor: extract expense editability helper`
- Dengan ini, pola `expense` menjadi jauh lebih sejajar dengan `another_income`.


### Helper / perapihan yang sudah dilakukan
Di `services/expense_service.go`:
- `ParseExpenseDate(...)`
- `NormalizeExpensePayment(...)`

Di `expense_handler.go`:
- parsing tanggal dipisah
- normalisasi payment update dipisah agar nilai kosong tidak overwrite payment lama

### Hasil
- `POST /api/expenses` tervalidasi runtime
- `PUT /api/expenses/:id` tervalidasi runtime
- bug lama terkait overwrite payment kosong tetap terjaga agar tidak kembali muncul

## 12. Another Income

Another income sekarang juga sudah mulai masuk ke pola refactor ringan yang sama dengan expense, bahkan sedikit lebih rapi.

### Helper / perapihan yang sudah dilakukan
Di `services/another_income_service.go`:
- `ParseAnotherIncomeDate(...)`
- `NormalizeAnotherIncomePayment(...)`
- `EnsureAnotherIncomeEditable(...)`
- `ErrAnotherIncomeDataExpiredToEdit`

Di `another_income_handler.go`:
- parsing tanggal dipisah
- normalisasi payment update dipisah
- editability check dipisah ke helper reusable

### Hasil
- `POST /api/another-incomes` tervalidasi runtime
- `PUT /api/another-incomes/:id` tervalidasi runtime
- `DELETE /api/another-incomes/:id` tervalidasi runtime
- another income sekarang cukup sehat dan kualitasnya sudah seimbang dengan expense untuk fase ini


## 13. User Branch

User branch sekarang sudah mulai disentuh sebagai modul non-transaksional strategis lanjutan.

### Refactor kecil yang sudah dilakukan
Di `services/user_branch_service.go`:
- `BuildUserBranchRows(...)`
- `UserBranchExists(...)`
- `BuildUserDetailBranches(...)`

Di `user_branch_handler.go`:
- formatting response list dipisah ke helper
- duplicate check create dipisah ke helper
- formatter detail branch untuk endpoint detail user dipisah ke helper

### Runtime verification
- `GET /api/user-branches` -> 200
- `POST /api/user-branches` -> 409 expected untuk pasangan user-branch yang memang sudah ada
- `GET /api/detail-users/USR250118132201` -> 200

### Catatan penting
- Sempat ada false start saat helper service memakai referensi type model yang salah dan membuat build pecah.
- Sudah langsung direpair dengan commit `fix: restore user branch service build`.
- Status akhir modul user_branch untuk fase ini: build kembali hijau dan endpoint utama yang diuji sudah sehat.


## 14. Dashboard

Dashboard disentuh ringan lagi sebagai modul non-transaksional strategis dengan risiko rendah.

### Refactor kecil yang sudah dilakukan
Di `services/dashboard_service.go`:
- `CalculateProfitPercentages(...)`
- `BuildDailyProfitByUserReportData(...)`

Di `dashboard_handler.go`:
- kalkulasi HPP/profit percentage mingguan dipisah ke helper
- formatter data profit harian per user dipisah ke helper

### Runtime verification
- `GET /api/dashboard/weekly-profit-report` -> 200
- `GET /api/dashboard/daily-profit-report` -> 200
- `GET /api/dashboard/profit-today-by-user` -> 200

### Penilaian fase ini
- Dashboard mendapat sentuhan refactor ringan yang aman.
- Cost-benefit untuk pembongkaran lebih dalam masih kecil, jadi cukup sehat untuk fase saat ini.

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
- kualitasnya sekarang sudah sejajar dengan modul transaksi utama lain untuk fase ini

### Sale return
- kualitasnya sekarang sudah sejajar dengan modul transaksi utama lain untuk fase ini

### First stock
- kualitasnya sekarang sudah sejajar dengan modul transaksi utama lain untuk fase ini

### Modul lain di luar transaksi utama
Mulai disentuh:
- opname sudah masuk fase refactor sehat

### Modul yang masih bisa diperdalam berikutnya
- dashboard (jika suatu saat perlu cleanup ringan)
- daily asset (jika ingin diteruskan lebih dalam)
- expense / another income (jika ingin dibuat satu tingkat lebih seragam lagi)
- modul non-transaksional lain yang masih belum tersentuh

## Rekomendasi tahap selanjutnya

Urutan yang disarankan setelah checkpoint ini:
1. tentukan modul transaksi berikutnya yang mau diperdalam
2. lanjutkan dengan pola yang sama, kecil tapi konsisten
3. pertimbangkan kapan mulai mengekstrak helper lokal menjadi service/usecase yang lebih resmi

### Rekomendasi prioritas
Karena modul transaksi utama sekarang sudah relatif merata kualitas refactornya, pilihan paling masuk akal berikutnya adalah:
- lanjut ke fitur/menu lain di luar transaksi utama
- atau mulai cleanup / normalisasi layer jika memang diperlukan

Jika ingin melanjutkan coverage, kandidat berikutnya bisa berupa:
- modul non-transaksional lain yang masih besar dan belum tersentuh
- atau pendalaman tambahan pada cashflow ringan jika ingin kualitasnya lebih simetris

## Kesimpulan

Fase ini berhasil menghasilkan tiga hal penting:
1. runtime baseline yang kuat
2. pola refactor incremental yang terbukti aman
3. hampir seluruh modul transaksi utama yang sudah disentuh kini mencapai kualitas refactor yang lebih merata: purchase, sale, duplicate receipt, buy return, sale return, dan first stock

Purchase tetap menjadi template refactor paling matang.
Sale dan duplicate receipt sudah mengikuti pola yang stabil.
Buy return, sale return, dan first stock kini sudah berhasil didorong ke level kualitas yang sebanding untuk fase ini.
