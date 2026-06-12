# 🏥 Apotek-Clean

> Refactor bertahap dari repository **fiber-apotek** menuju struktur yang lebih rapi dan lebih dekat ke **Clean Architecture**, tanpa gegabah merusak perilaku API yang sudah dipakai.

## ✨ Gambaran Singkat

Project ini fokus pada 2 hal utama:

- **menjaga kompatibilitas API lama** agar frontend lama tetap aman
- **membersihkan struktur kode** modul demi modul supaya lebih mudah dirawat, dites, dan dikembangkan

Saat ini repo sudah berada di fase yang **cukup matang untuk clone, configure, build, run, dan konsumsi endpoint aktif utama** yang sudah diaudit runtime.

---

## 📌 Status Saat Ini

### ✅ Yang sudah tervalidasi
- fresh clone berhasil **build**
- binary hasil build berhasil **run**
- auth **2 tahap** berhasil
- endpoint protected utama berhasil diakses
- mayoritas master data inti tervalidasi
- transaksi inti tervalidasi
- export utama dan banyak export item-level tervalidasi
- dashboard dan reporting dasar tervalidasi

### ⚠️ Yang perlu tetap diingat
- belum fair kalau diklaim **100% identik** dengan repo lama di semua sudut perilaku
- ada beberapa **legacy exception** yang memang sengaja dipertahankan
- ada beberapa **known behavior** yang harus dibaca sebagai kontrak aktif, bukan bug refactor

📄 Dokumen status paling penting:
- `RUNTIME_AUDIT.md` → baseline audit runtime
- `REFACTOR_CHECKPOINT.md` → checkpoint refactor terbaru
- `LEGACY_REPLACEMENT_CHECKLIST.md` → checklist apakah repo baru sudah cukup aman menggantikan repo lama
- `TEST_COVERAGE_MAP.md` → peta coverage smoke, regression, dan deeper verification saat ini
- `API_NORMALIZATION_PLAN.md` → strategi normalisasi API non-breaking
- `API_CONTRACT_DIFF.md` → catatan old vs new contract
- `POSTMAN_AUDIT_NOTES.md` → audit endpoint terhadap koleksi lama

---

## 🎯 Tujuan Refactor

Project ini **bukan** rewrite besar sekali jalan.

Pendekatan yang dipakai:

1. **pertahankan behavior API dulu**
2. **pecah bagian padat sedikit demi sedikit**
3. **build + smoke test setelah langkah aman**
4. **commit kecil langsung ke `main` bila sudah stabil**

Kenapa begitu?
Karena target utamanya bukan sekadar kode lebih cantik, tapi **repo baru benar-benar bisa menggantikan repo lama secara aman**.

---

## 🧱 Struktur Project

```text
apotek-clean/
├── cmd/
│   └── app/
│       └── main.go
├── configs/
├── helpers/
├── internal/
│   └── adapters/
│       └── driving/
│           └── http/
│               ├── handlers/
│               └── routes/
├── middlewares/
├── models/
├── services/
├── seeders/
├── schedulers/
├── controllers/                    # sisa transisi legacy
├── routes/                         # sisa transisi legacy
├── menus.json
├── RUNTIME_AUDIT.md
├── REFACTOR_CHECKPOINT.md
├── LEGACY_REPLACEMENT_CHECKLIST.md
└── README.md
```

### Poin penting struktur saat ini
- **entry point resmi** aplikasi: `cmd/app/main.go`
- route utama sekarang dipusatkan di:
  - `internal/adapters/driving/http/routes`
- handler aktif utama sekarang dipusatkan di:
  - `internal/adapters/driving/http/handlers`
- struktur legacy masih ada sebagian, tapi sudah makin dipersempit dari jalur aktif runtime

---

## ✅ Modul yang Sudah Stabil di Runtime

Sudah tervalidasi melalui smoke test runtime:

- 🔐 auth dua tahap
- 🏢 branches
- 👤 users
- 📦 master data utama
- 🛒 purchase
- 💳 sale
- ↩️ buy return
- 🔁 sale return
- 🧾 duplicate receipt
- 📥 first stock
- 🧮 opname basic
- 📊 reporting dasar
- 📄 export Excel/PDF utama
- 🧩 export item-level
- 📈 export audit / report / dashboard

---

## 🚀 Quick Start

### 1) Clone repository

```bash
git clone <repo-url>
cd apotek-clean
```

### 2) Siapkan environment file

```bash
cp .example_env .env
```

Lalu sesuaikan `.env` dengan environment Anda.

### 3) Install dependency

```bash
go mod tidy
```

> Opsional, helper command juga tersedia lewat `Makefile`.

### 4) Jalankan aplikasi

#### Opsi A, langsung dengan Go

```bash
go run ./cmd/app/main.go
```

#### Opsi B, build binary lalu jalankan

```bash
go build -o ./bin/apotek ./cmd/app
./bin/apotek
```

---

## ⚙️ Konfigurasi Environment

Gunakan file `.env` di root project.

### Contoh minimal env yang sesuai runtime saat ini

```env
PORT=9002
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=postgres
DB_PASS=postgres
DB_NAME=apotek_clean
JWT_SECRET_KEY=change-me
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_PASS=
REDIS_SHORT=0
PROJECT_NAME=core-dev
APPNAME=Dev.Core
```

### Catatan penting env

Beberapa nama env **masih mengikuti kode lama**, jadi perhatikan ini:

- ✅ `DB_PASS`
- ❌ **bukan** `DB_PASSWORD`

- ✅ `REDIS_HOST`
- ✅ `REDIS_PORT`
- ✅ `REDIS_PASS`
- ✅ `REDIS_SHORT`
- ❌ **bukan** `REDIS_ADDR`

- ✅ `JWT_SECRET_KEY`

### Path resolution yang sudah dihardening

Sekarang aplikasi akan mencari file penting dari project root dengan lebih aman, jadi tidak mudah rusak hanya karena beda working directory.

Yang sudah diamankan:
- `.env`
- `menus.json`
- beberapa file project path penting lain

Artinya skenario ini sekarang sama-sama didukung:
- menjalankan app dari root repo
- menjalankan binary dari folder `bin/`

---

## 🔐 Auth Flow yang Dipakai Runtime

Project ini memakai **auth 2 tahap**.

### Tahap 1
`POST /api/login`

Hasilnya token awal, tapi token ini **belum cukup** untuk semua endpoint branch-scoped.

### Tahap 2
1. `GET /api/list_branches`
2. `POST /api/set_branch`

Baru setelah itu Anda dapat token branch-scoped yang dipakai untuk endpoint protected utama.

---

## 🧪 Build dan Verifikasi

### Build utama

```bash
go build -o ./bin/apotek ./cmd/app
```

Atau lewat Makefile:

```bash
make build
```

### Unit / package build check

```bash
go test ./...
```

### Runtime smoke test yang biasa dipakai

- login
- list branches
- set branch
- akses endpoint protected utama
- smoke test transaksi penting
- cek stok naik/turun/rollback
- cek report
- cek export

### Helper scripts yang sekarang tersedia

#### Restart lokal cepat
```bash
./scripts/restart_local.sh
```

Opsional ganti port:

```bash
PORT=9017 ./scripts/restart_local.sh
```

#### Smoke regression baseline berbasis inventory
```bash
./scripts/regression_inventory_smoke.py
```

Daftar endpoint baseline-nya disimpan di:
- `scripts/regression_cases.json`

Opsional ganti base URL / sample ID:

```bash
BASE_URL=http://127.0.0.1:9017 \
SMOKE_PURCHASE_ID=PUR49589903CJ0G \
SMOKE_SALE_ID=SAL466951VV0DS5 \
./scripts/regression_inventory_smoke.py
```

#### One command restart + smoke
```bash
./scripts/run_local_regression.sh
```

Opsional ganti port:

```bash
PORT=9017 ./scripts/run_local_regression.sh
```

#### Fresh clone smoke
```bash
./scripts/fresh_clone_smoke.sh
```

Opsional ganti port / tmp dir:

```bash
PORT=9019 TMP_DIR=/tmp/apotek-clean-fresh-smoke ./scripts/fresh_clone_smoke.sh
```

#### Shortcut via Makefile
```bash
make restart PORT=9017
make smoke PORT=9017
make smoke-fresh PORT=9019
make test
```

Untuk detail hasil validasi, lihat:
- `RUNTIME_AUDIT.md`
- `REFACTOR_CHECKPOINT.md`
- `LEGACY_REPLACEMENT_CHECKLIST.md`

---

## 🧩 Kompatibilitas dengan Repo Lama

Secara praktis, repo ini dibangun supaya **cukup aman menggantikan repo lama** untuk alur operasional utama.

### Yang sudah kuat
- clone → configure → build → run
- auth 2 tahap
- konsumsi endpoint aktif utama
- banyak kontrak utama tetap kompatibel
- beberapa alias compatibility sudah ditambahkan untuk mengurangi risiko break di frontend

### Legacy exception yang sengaja dipertahankan
Area berikut **tidak dipaksa dinormalisasi** karena diperlakukan sebagai kontrak legacy yang tetap hidup:

- `opname`
- `opname-item`

Contoh penting:
- `GET /api/opname-items/` tetap mengikuti pola legacy dengan body JSON:

```json
{
  "opname_id": "..."
}
```

### Known behavior yang harus dipahami
- area `returns` tidak boleh diasumsikan CRUD penuh seperti transaksi lain
- keputusan efek bisnis return dibedakan tegas:
  - `buy_return` = koreksi stok + transaction report, tidak ikut omset / margin penjualan
  - `sale_return` = koreksi stok + koreksi omset / margin harian
  - `duplicate receipt` = transaksi penjualan penuh, jadi menyentuh stok + omset + margin
- beberapa sample ID di koleksi lama memang sudah usang, jadi bisa terlihat seolah route bermasalah padahal sample-nya yang tidak valid
- false alarm runtime bisa muncul kalau port `:9002` masih dipegang binary lama `(deleted)`

Kalau butuh ringkasan operasional paling jujur, baca:
- `LEGACY_REPLACEMENT_CHECKLIST.md`

---

## 🛠️ Progress Refactor Modul

### 🛒 Purchase
Modul paling jauh direfactor saat ini.
Sudah dirapikan bertahap untuk:
- editability helper
- total helper
- date parsing helper
- item value preparation
- dependency lookup
- product update builder
- response builder
- supplier lookup
- item response builder
- item model builder
- item preparation helper lokal
- transaction orchestration helper lokal
- rollback response helper

### 💳 Sale
Sudah mengikuti pola purchase dengan perapihan pada:
- editability helper
- totals calculation helper
- rollback response helper
- stock update helper
- product lookup + stock validation
- sale item preparation
- orchestration helper

### 🧾 Duplicate Receipt
Sudah mengikuti pola yang sama pada area:
- editability helper
- totals helper
- product lookup
- stock validation
- item preparation
- wiring handler ke helper
- orchestration helper

---

## 🗺️ Roadmap Terdekat

Prioritas berikutnya:

- menutup known gap yang masih realistis dirapikan
- terus mengurangi ketergantungan ke struktur legacy
- menjaga supaya surface API tetap frontend-friendly
- memperkuat status repo baru sebagai pengganti operasional repo lama

---

## 📎 Catatan Tambahan

- fitur yang bergantung pada Google Drive dapat memerlukan `credentials.json`
- file tersebut **bukan blocker** untuk startup inti aplikasi, login, dan akses endpoint dasar
- refactor dilakukan incremental, bukan rewrite besar sekali jalan
- preferensi kerja repo ini adalah **lebih baik commit kecil yang lolos build dan smoke test** daripada refactor besar yang rawan regression

---

## 📚 Dokumen Pendukung

- `RUNTIME_AUDIT.md`
- `REFACTOR_CHECKPOINT.md`
- `LEGACY_REPLACEMENT_CHECKLIST.md`
- `TEST_COVERAGE_MAP.md`
- `POSTMAN_AUDIT_NOTES.md`
- `API_NORMALIZATION_PLAN.md`
- `API_CONTRACT_DIFF.md`

---

## 📄 Lisensi

Mengikuti lisensi dari project sumber jika tidak ditentukan lain.
