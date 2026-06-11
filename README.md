# Apotek-Clean

Refactor repository **fiber-apotek** ke arah **Clean Architecture** secara bertahap, dengan prioritas utama menjaga perilaku API tetap stabil sambil membersihkan struktur kode modul demi modul.

## Status Saat Ini

Project ini sudah berada pada fase:
- runtime utama tervalidasi
- route internal utama aktif
- export utama sudah aman
- modul transaksi besar mulai direfactor bertahap

Dokumen progres penting:
- `RUNTIME_AUDIT.md` → baseline audit runtime
- `REFACTOR_CHECKPOINT.md` → checkpoint refactor terbaru

## Fokus Refactor

Tujuan fase saat ini:
- menjaga kompatibilitas API lama
- memindahkan alur HTTP utama ke `internal/adapters/driving/http`
- mengurangi ketergantungan ke struktur legacy secara bertahap
- merapikan handler besar menjadi helper/service kecil yang lebih aman dirawat

## Struktur Project Saat Ini

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
├── controllers/        # sisa transisi legacy yang belum seluruhnya dihapus
├── routes/             # sisa transisi legacy yang belum seluruhnya dihapus
├── RUNTIME_AUDIT.md
├── REFACTOR_CHECKPOINT.md
└── README.md
```

## Modul yang Sudah Stabil di Runtime

Sudah tervalidasi melalui smoke test runtime:
- auth dua tahap
- branches
- users
- master data utama
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

## Progress Refactor Modul

### Purchase
Paling jauh direfactor saat ini.
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

### Sale
Sudah mulai mengikuti pola purchase.
Sudah dirapikan untuk:
- editability helper
- totals calculation helper
- rollback response helper
- stock update helper
- product lookup + stock validation
- sale item preparation
- orchestration helper

### Duplicate Receipt
Sudah mulai mengikuti pola yang sama.
Sudah dirapikan untuk:
- editability helper
- totals helper
- product lookup
- stock validation
- item preparation
- wiring handler ke helper
- orchestration helper

## Konfigurasi Environment

Gunakan file `.env` di root project.

Langkah awal paling aman:

```bash
cp .example_env .env
```

Lalu sesuaikan isi `.env` dengan koneksi database, Redis, dan secret milik environment Anda.

Catatan penting:
- proses aplikasi akan mencari `.env` di working directory aktif, folder binary, lalu parent folder binary
- skenario umum `./bin/apotek` dari root repo maupun menjalankan binary dari folder `bin/` sekarang sama-sama didukung

### Contoh env yang relevan dengan kode saat ini

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
```

Perhatikan bahwa beberapa nama env masih mengikuti kode lama, misalnya:
- `DB_PASS` bukan `DB_PASSWORD`
- `JWT_SECRET_KEY` masih dipakai

## Menjalankan Project

### Install dependency

```bash
go mod tidy
```

### Jalankan aplikasi

```bash
go run ./cmd/app/main.go
```

### Build binary

```bash
go build -o ./bin/apotek ./cmd/app
./bin/apotek
```

Port dibaca dengan urutan fallback:
1. `APP_PORT`
2. `PORT`
3. `SERVER_PORT`
4. default `9001`

## Build dan Verifikasi

Build utama yang dipakai selama fase ini:

```bash
go build -o ./bin/apotek ./cmd/app
```

Verifikasi runtime dilakukan bertahap dengan:
- login
- set branch
- akses endpoint penting
- smoke test transaksi utama
- cek efek stok
- cek report
- cek export

## Catatan Penting

- Struktur legacy belum seluruhnya hilang, tetapi bagian aktifnya sudah jauh berkurang
- Route utama sekarang dipusatkan melalui `internal/adapters/driving/http/routes`
- Entry point resmi aplikasi adalah `cmd/app/main.go`
- Refactor dilakukan incremental, tidak one-shot besar, untuk meminimalkan regression
- Commit kecil yang lolos build dan smoke test lebih diutamakan dibanding refactor besar sekali jalan

## Roadmap Fase Berikutnya

Prioritas terdekat:
- lanjut merapikan modul transaksi lain dengan pola yang sama
- memperdalam cleanup sale / duplicate receipt bila masih ada area padat
- lanjut ke modul seperti first stock, buy return, atau sale return
- secara bertahap mengurangi sisa ketergantungan ke struktur legacy

## Testing

### Unit / package build check

```bash
go test ./...
```

### Runtime smoke test

Lihat ringkasan hasil validasi di:
- `RUNTIME_AUDIT.md`
- `REFACTOR_CHECKPOINT.md`

## Lisensi

Mengikuti lisensi dari project sumber jika tidak ditentukan lain.
