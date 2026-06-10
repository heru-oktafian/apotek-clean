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
