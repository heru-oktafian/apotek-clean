# 🧪 Test Coverage Map

Tanggal pembaruan: 2026-06-12
Target dokumen ini: memetakan coverage pengujian repo `apotek-clean` agar jelas mana yang sudah dijaga oleh smoke/regression automation, mana yang masih butuh verifikasi workflow bisnis lebih dalam.

---

## 1. Cara baca dokumen ini

Di repo ini kita memakai 3 lapis validasi utama:

### A. Smoke test
Tujuan: menjawab pertanyaan **"app masih hidup atau sudah jebol?"**

Ciri:
- cepat
- fokus ke jalur paling penting
- dipakai setelah build, restart, atau patch kecil

### B. Regression test
Tujuan: menjawab pertanyaan **"yang tadinya sudah sehat, apakah masih sehat setelah ada perubahan baru?"**

Ciri:
- lebih luas dari smoke
- cocok untuk menjaga repo refactor agar tidak balik rusak
- sangat penting untuk perubahan handler, route, helper, atau startup path

### C. Full / deeper verification
Tujuan: menjawab pertanyaan **"workflow bisnis di dalam endpoint benar atau tidak?"**

Ciri:
- lebih mahal dan lebih lambat
- tidak cukup hanya cek status `200`
- perlu cek stok, total/subtotal, rollback, report/profit, dan konsistensi data

---

## 2. Ringkasan status coverage saat ini

### ✅ Sudah ada automation yang bisa dijalankan ulang
- startup lokal
- fresh clone build + run
- auth 2 tahap
- baseline endpoint JSON penting
- baseline export penting
- empty-state return support
- mutation ringan `expense` dan `another-income`
- mutation top-down master data batch awal:
  - `member-categories` create/detail/update/delete
  - `members` create/detail/update/delete
  - `product-categories` create/detail/update/delete
  - `units` create/detail/update/delete
  - `products` create/detail/update/delete
  - `supplier-categories` create/detail/update/delete
  - `suppliers` create/detail/update/delete
- deep stock check `first_stock` (stok naik saat create, rollback saat delete)

### ⚠️ Masih butuh verifikasi manual / deeper check
- stok sebelum/sesudah pada semua transaksi inti
- rollback kompleks semua transaksi
- profit/report sinkron untuk semua mutasi
- seluruh endpoint inventory sampai detail workflow internal
- semua variasi payload error/edge case

---

## 3. Coverage map per level

## A. Smoke coverage

### A1. Startup & clone-run
Status: **terautomasi**

Cakupan:
- fresh clone repo
- copy `.env`
- build `./cmd/app`
- jalankan binary hasil build
- login dasar
- akses endpoint penting setelah startup

Script terkait:
- `scripts/fresh_clone_smoke.sh`
- `scripts/restart_local.sh`
- `Makefile` target `smoke-fresh`

Command:
```bash
./scripts/fresh_clone_smoke.sh
make smoke-fresh PORT=9019
```

### A2. Runtime smoke lokal
Status: **terautomasi**

Cakupan:
- restart app lokal
- auth dua tahap
- smoke endpoint baseline

Script terkait:
- `scripts/run_local_regression.sh`
- `scripts/restart_local.sh`
- `Makefile` target `restart`

Command:
```bash
./scripts/run_local_regression.sh
make restart PORT=9017
```

---

## B. Regression coverage

### B1. Baseline endpoint regression
Status: **terautomasi**

Sumber case:
- `scripts/regression_cases.json`

Runner:
- `scripts/regression_inventory_smoke.py`
- `Makefile` target `smoke`

Command:
```bash
./scripts/regression_inventory_smoke.py
make smoke BASE_URL=http://127.0.0.1:9002
```

### B2. Group coverage saat ini

#### Group: `json_get_cases`
Status: **terautomasi**

Endpoint baseline yang dicek:
- `/api/profile`
- `/api/expenses`
- `/api/another-incomes`
- `/api/first-stocks`
- `/api/buy-returns`
- `/api/sale-returns`
- `/api/cmb-purchases`
- `/api/cmb-sales`
- `/api/dashboard/daily-profit-report`
- `/api/report/neraca-saldo`

Validasi yang dilakukan:
- status code
- content-type JSON

#### Group: `export_cases`
Status: **terautomasi**

Export yang dicek:
- expenses pdf/excel
- another-incomes pdf/excel
- first-stocks pdf/excel
- daily-assets excel

Validasi yang dilakukan:
- status code
- content-type file yang sesuai

#### Group: `return_support_cases`
Status: **terautomasi**

Cakupan:
- combo buy return
- combo sale return
- list return bulan kosong
- combo purchase/sale bulan kosong
- combo return dengan sample ID kosong

Validasi yang dilakukan:
- status code
- content-type JSON
- `data` harus berbentuk **list** (`[]` saat kosong, bukan `null`)

### B3. Mutation regression ringan
Status: **terautomasi**

Cakupan:
- `POST /api/expenses`
- `PUT /api/expenses/:id`
- `DELETE /api/expenses/:id`
- `POST /api/another-incomes`
- `PUT /api/another-incomes/:id`
- `DELETE /api/another-incomes/:id`

Tujuan:
- memastikan perubahan refactor tidak merusak create/update/delete ringan yang sering dipakai sebagai baseline kesehatan transaksi non-kompleks

Catatan:
- mutation ini memakai fixture sementara dan dibersihkan dalam flow yang sama

### B4. Mutation regression top-down master data
Status: **terautomasi**

Cakupan batch awal sesuai urutan inventory yang aman:
- `member-categories` create/detail/update/delete
- `members` create/detail/update/delete
- `product-categories` create/detail/update/delete
- `units` create/detail/update/delete
- `products` create/detail/update/delete
- `supplier-categories` create/detail/update/delete
- `suppliers` create/detail/update/delete

Validasi yang dilakukan:
- create fixture sementara
- ambil detail resource hasil create
- update resource
- delete resource
- cleanup fallback bila flow terputus di tengah

Nilai tambah:
- coverage automation sekarang lebih dekat ke inventory top-down nyata, bukan hanya read-only + transaksi ringan
- blok master data atas yang sebelumnya dominan manual kini sudah punya pagar regresi formal

---

## C. Full / deeper verification coverage

Status: **belum seluruhnya terautomasi**

Jenis verifikasi ini saat ini lebih banyak terdokumentasi lewat audit runtime dan inventory, belum semuanya menjadi script otomatis.

### C1. Workflow stok
Status: **parsial, mulai masuk automation formal**

Sudah pernah dibuktikan:
- purchase menambah stok
- sale mengurangi stok
- buy return mengurangi stok
- sale return menambah stok
- first stock menambah stok
- duplicate receipt mengurangi stok
- delete sale / duplicate receipt melakukan rollback stok

Yang sekarang sudah ikut terautomasi:
- `first_stock` create menaikkan stok produk uji
- `first_stock` delete mengembalikan stok ke nilai awal

Sumber acuan:
- `RUNTIME_AUDIT.md`
- `API_ENDPOINT_INVENTORY.xlsx`

### C2. Rollback dan cleanup
Status: **parsial**

Sudah pernah dibuktikan pada beberapa flow utama, tetapi belum semuanya menjadi automation formal.

Cakupan yang masih layak didorong:
- rollback transaksi inti lebih konsisten dalam script
- cleanup fixture untuk flow yang lebih kompleks
- pengujian return flow yang lebih kaya tanpa mengubah kontrak frontend

### C3. Report / profit / omzet sinkronisasi
Status: **parsial**

Sudah pernah disentuh di audit manual, tetapi belum menjadi assertion otomatis yang luas.

Yang masih ideal dikerjakan nanti:
- setelah create sale, cek profit/report terkait
- setelah delete/rollback, cek report ikut pulih
- cek konsistensi report harian/bulanan setelah mutasi tertentu

---

## 4. Coverage map per domain

## Auth & startup
- Smoke: **ya**
- Regression: **ya**
- Full/deeper: **cukup** untuk baseline operasional

## Master data inti
- Smoke: **sebagian besar ya**
- Regression: **belum luas**
- Full/deeper: **masih banyak manual via inventory**

## Expenses / another-incomes
- Smoke: **ya**
- Regression: **ya**, termasuk mutation ringan
- Full/deeper: **cukup untuk baseline, belum banyak edge case**

## Returns
- Smoke: **ya**
- Regression: **ya**, terutama support route dan empty-state
- Full/deeper: **masih perlu hati-hati**, karena kontraknya memang tidak selalu CRUD penuh

## First stock
- Smoke: **ya**
- Regression: **ya** untuk list/export baseline
- Full/deeper: **mulai terautomasi** untuk stok naik saat create dan rollback saat delete

## Sale / Purchase / Duplicate Receipt
- Smoke: **belum dijadikan mutation otomatis penuh**
- Regression: **baru baseline support/list tertentu**
- Full/deeper: **sudah banyak diverifikasi manual**, tetapi belum seluruhnya masuk runner otomatis

## Opname / opname-item
- Smoke: **minimal**
- Regression: **belum diprioritaskan**
- Full/deeper: **tetap dianggap legacy exception**, jangan dipaksa seragam dengan domain lain

---

## 5. Alat yang tersedia sekarang

### Scripts
- `scripts/restart_local.sh`
- `scripts/run_local_regression.sh`
- `scripts/fresh_clone_smoke.sh`
- `scripts/regression_inventory_smoke.py`
- `scripts/regression_cases.json`

### Makefile targets
- `make tidy`
- `make build`
- `make run`
- `make restart`
- `make smoke`
- `make smoke-fresh`
- `make test`

---

## 6. Gap terbesar yang masih tersisa

Urutan paling realistis tanpa mengubah habit frontend:

### 1. Perluas `regression_cases.json`
Supaya coverage otomatis makin dekat ke inventory nyata, bukan cuma baseline inti.

### 2. Tambah mutation automation untuk transaksi inti terpilih
Prioritas aman:
- sale ringan
- purchase ringan
- first stock ringan

### 3. Tambah assertion workflow internal
Yang paling bernilai:
- stok sebelum/sesudah
- rollback setelah delete
- report / profit sync pada flow yang dipilih

### 4. Pisahkan level test dengan lebih tegas
Contoh target ke depan:
- `smoke` = cepat dan ringan
- `regression` = endpoint baseline + export + mutation ringan
- `deep` = stok / rollback / report assertion

---

## 7. Kesimpulan

Kalau diringkas:

- repo ini sekarang **sudah punya smoke + regression baseline yang nyata dan bisa diulang**
- coverage itu sudah cukup kuat untuk menjaga agar refactor harian tidak gampang bikin area yang sehat balik rusak
- tetapi **belum semua workflow bisnis dalam** sudah dijadikan automation formal

Kalimat paling jujur untuk status saat ini:

> Smoke dan regression baseline sudah cukup matang untuk jadi pagar harian refactor.
>
> Full business verification masih harus terus dibangun bertahap, terutama pada stok, rollback, dan sinkronisasi report.
