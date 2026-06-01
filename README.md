# Apotek-Clean

Refactor repository **fiber-apotek** ke struktur **Clean Architecture** agar maintenance, testing, dan pengembangan fitur berikutnya lebih rapi.

## Tujuan

- Memisahkan domain, use case, adapter, dan framework
- Menjaga perilaku API tetap sama seperti versi sebelumnya
- Memudahkan pengujian, penggantian infrastruktur, dan scaling project

## Struktur Proyek

```text
apotek-clean/
├── cmd/
│   └── app/
│       └── main.go
├── configs/
│   ├── database_config.go
│   └── timezone_config.go
├── internal/
│   ├── core/
│   │   ├── entities/
│   │   ├── ports/
│   │   │   ├── driven/
│   │   │   └── driving/
│   │   └── usecases/
│   ├── adapters/
│   │   ├── driven/
│   │   │   └── postgres/
│   │   └── driving/
│   │       └── http/
│   │           ├── handlers/
│   │           └── routes/
│   ├── frameworks/
│   │   ├── auth/
│   │   ├── database/
│   │   ├── export/
│   │   │   ├── excel/
│   │   │   └── pdf/
│   │   ├── scheduler/
│   │   └── web/
│   └── utils/
├── controllers/
├── helpers/
├── middlewares/
├── models/
├── routes/
├── schedulers/
├── seeders/
├── services/
├── .env.example
├── go.mod
└── README.md
```

## Keterangan Layer

### 1. Core
Berisi logika inti aplikasi.

- `entities/` untuk model domain utama
- `usecases/` untuk aturan bisnis
- `ports/` untuk kontrak interface antara core dan layer luar

### 2. Adapters
Berisi penghubung antara core dan dunia luar.

- `driving/http/handlers/` untuk handler HTTP
- `driving/http/routes/` untuk route registration
- `driven/postgres/` untuk implementasi repository PostgreSQL

### 3. Frameworks
Berisi detail teknis/infrastruktur.

- `web/` untuk setup Fiber
- `database/` untuk koneksi database
- `auth/` untuk JWT atau auth helper
- `scheduler/` untuk cron/scheduler
- `export/` untuk export Excel/PDF

## Konfigurasi Environment

Gunakan file `.env` di root project.

### Contoh `.env`

```env
APP_PORT=9001
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=apotek_clean
DB_SSLMODE=disable
JWT_SECRET=change-me
```

Jika belum ada, copy dari contoh:

```bash
cp .env.example .env
```

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
go build -o bin/apotek ./cmd/app
./bin/apotek
```

Server akan berjalan sesuai `APP_PORT` di `.env`, saat ini disiapkan untuk **9001**.

## Database

Project menggunakan PostgreSQL. Pastikan database yang sesuai dengan `.env` sudah tersedia.

Contoh quick check:

```bash
psql -h 127.0.0.1 -U postgres -d apotek_clean
```

## Testing

### Unit test

```bash
go test ./...
```

### API test

Gunakan Postman collection yang sudah disiapkan untuk memastikan seluruh endpoint tetap kompatibel dengan versi sebelumnya.

## Catatan Penting

- Struktur lama seperti `controllers/`, `services/`, `routes/`, dan `models/` masih ada selama masa transisi refactor
- Target akhirnya adalah seluruh alur utama berpindah ke `internal/`
- File `cmd/app/main.go` adalah entry point utama aplikasi baru

## Roadmap Refactor

- [x] Menyiapkan struktur Clean Architecture
- [x] Menyiapkan entities awal
- [x] Menyiapkan use case, ports, adapters, dan frameworks dasar
- [x] Menambahkan dukungan `.env`
- [ ] Merapikan dependency injection penuh di `cmd/app/main.go`
- [ ] Menyelesaikan seluruh implementasi repository dan handler sampai build benar-benar hijau
- [ ] Verifikasi endpoint via Postman collection
- [ ] Finalisasi dokumentasi deployment

## Deploy Sederhana

Contoh Dockerfile minimal:

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o /apotek ./cmd/app

FROM alpine:latest
WORKDIR /app
COPY --from=builder /apotek /apotek
COPY .env .
EXPOSE 9001
ENTRYPOINT ["/apotek"]
```

## Lisensi

Mengikuti lisensi dari project sumber jika tidak ditentukan lain.
