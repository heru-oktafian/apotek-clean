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
