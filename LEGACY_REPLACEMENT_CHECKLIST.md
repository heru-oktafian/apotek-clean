# Legacy Replacement Checklist

Tanggal pembaruan: 2026-06-12
Target: menilai apakah repo baru `apotek-clean` sudah cukup aman untuk menggantikan repo lama pada skenario clone -> configure -> build -> run -> konsumsi API.

## Putusan singkat

**Status saat ini: hampir siap sebagai pengganti operasional repo lama, tetapi belum layak diklaim 100% identik di semua sudut perilaku.**

Terjemahan praktisnya:
- untuk alur clone, isi `.env`, build, run, auth 2 tahap, dan konsumsi endpoint aktif utama, repo baru **sudah tervalidasi kuat**
- untuk beberapa area legacy khusus dan kontrak yang memang tidak sepenuhnya CRUD penuh, tetap harus dibaca sebagai **known behavior**, bukan bug refactor baru

## Checklist pengganti repo lama

### A. Clone dan startup dasar
- [x] Repo baru bisa di-clone bersih
- [x] Fresh clone berhasil `go build -o ./bin/apotek ./cmd/app`
- [x] Binary hasil build bisa dijalankan
- [x] `.env` tidak lagi bergantung murni pada current working directory
- [x] `menus.json` tidak lagi bergantung murni pada current working directory
- [x] Jalur project file penting sudah dipusatkan lewat resolver project root

### B. Konfigurasi onboarding
- [x] `.example_env` sudah disesuaikan dengan nama env yang benar-benar dipakai runtime
- [x] README sudah menjelaskan urutan env dan cara start yang aman
- [x] Skenario `./bin/apotek` dari root repo tervalidasi
- [x] Skenario menjalankan binary dari folder `bin/` tervalidasi

### C. Auth dan endpoint dasar
- [x] `POST /api/login` tervalidasi `200`
- [x] `GET /api/list_branches` tervalidasi `200`
- [x] `POST /api/set_branch` tervalidasi `200`
- [x] `GET /api/profile` tervalidasi `200`
- [x] `GET /api/menus` tervalidasi `200`

### D. Branch-scoped smoke baseline
- [x] Endpoint branch-scoped contoh berhasil diakses setelah auth 2 tahap
- [x] Dari fresh clone, `GET /api/expenses?...` tervalidasi `200`

### E. Surface API aktif utama
- [x] Mayoritas master data inti sudah diverifikasi runtime
- [x] Transaksi inti sudah diverifikasi runtime
- [x] Export utama dan item-level utama sudah diverifikasi runtime
- [x] Dashboard/report dasar sudah diverifikasi runtime
- [x] Inventory workbook sudah disinkronkan jujur untuk endpoint yang sudah diuji

## Area yang dianggap aman untuk penggantian operasional

Area berikut sudah cukup kuat untuk diperlakukan sebagai baseline aktif repo baru:
- auth dua tahap
- branch/user/master data inti
- products / suppliers / categories / units / conversions
- purchases
- sales
- duplicate receipts
- expenses
- another incomes
- first stocks
- reporting dasar
- dashboard report utama
- export utama dan banyak export item-level

## Legacy exception yang sengaja dipertahankan

Ini **bukan** blocker, tetapi harus dipahami sebagai kontrak khusus yang memang tidak dipaksa dinormalisasi:
- seluruh area `opname`
- seluruh area `opname-item`
- contoh penting: `GET /api/opname-items/` tetap mengikuti kontrak legacy dengan body JSON `{ "opname_id": "..." }`

## Known behavior / known gap yang masih harus diingat

### 1. Return flow tidak boleh diasumsikan CRUD penuh seperti transaksi lain
- create/detail/export sudah tervalidasi
- beberapa percobaan `PUT` / `DELETE` root return memberi `405`
- beberapa route item return bukan pola `.../all/:id` seperti transaksi lain
- ini harus dibaca sebagai **kontrak nyata aktif**, bukan sekadar bug test

### 2. Tidak semua perilaku legacy identik 100%
Repo baru sudah sangat kompatibel untuk endpoint aktif yang diuji, tetapi belum boleh diklaim identik mutlak di semua sudut karena:
- ada route legacy khusus
- ada sample ID Postman lama yang sudah usang
- ada beberapa area yang sengaja dipertahankan apa adanya demi kompatibilitas, bukan dibersihkan total

### 3. `credentials.json` belum dijadikan fokus kesiapan clone-run
- saat ini sengaja diparkir
- relevan untuk fitur Google Drive, bukan untuk startup inti aplikasi
- bukan blocker utama untuk login, auth flow, dan konsumsi endpoint dasar

### 4. False alarm runtime masih mungkin terjadi bila listener lama belum dimatikan
- khususnya bila port `:9002` masih dipegang binary lama `(deleted)`
- sebelum menyimpulkan source rusak, selalu pastikan proses listener yang aktif memang binary terbaru

## Kesimpulan operasional

Kalimat yang paling jujur untuk status sekarang adalah:

> Repo baru sudah layak dipakai sebagai baseline pengganti operasional repo lama untuk skenario clone, konfigurasi env, build, run, auth 2 tahap, dan konsumsi endpoint aktif utama yang sudah diaudit.

Namun kalimat berikut **belum** aman diucapkan tanpa catatan:

> Semua endpoint dan semua perilaku legacy sudah 100% identik tanpa pengecualian.

## Rekomendasi penggunaan sekarang

Kalau mau mulai memakai repo baru sebagai default kerja:
1. clone repo baru
2. salin `.example_env` menjadi `.env`
3. sesuaikan `.env` ke environment target
4. build `./cmd/app`
5. jalankan app
6. validasi auth 2 tahap dan 2-3 endpoint branch-scoped penting di environment target

## Referensi dokumen pendukung
- `README.md`
- `RUNTIME_AUDIT.md`
- `REFACTOR_CHECKPOINT.md`
- `POSTMAN_AUDIT_NOTES.md`
- `API_NORMALIZATION_PLAN.md`
- `API_CONTRACT_DIFF.md`
