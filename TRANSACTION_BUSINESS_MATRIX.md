# 📊 Transaction Business Matrix

Tanggal pembaruan: 2026-06-12

Dokumen ini merangkum **efek bisnis utama per domain transaksi** agar refactor berikutnya tidak rancu antara:
- pergerakan stok
- omset penjualan
- profit / margin
- transaction report / neraca
- daily profit report

Tujuan utamanya: menjaga agar perubahan kode mengikuti **makna bisnis transaksi**, bukan hanya pola route atau bentuk CRUD.

---

## 1. Aturan baca

### Arti kolom
- **Stok**: apakah transaksi mengubah kuantitas barang di persediaan.
- **Omset**: apakah transaksi mempengaruhi angka penjualan (`total_sales`) di `daily_profit_reports`.
- **Profit / margin**: apakah transaksi mempengaruhi `profit_estimate` di `daily_profit_reports`.
- **Transaction report**: apakah transaksi masuk ke `transaction_reports` / neraca saldo.
- **Daily profit report**: apakah transaksi disinkronkan ke `daily_profit_reports`.

### Nilai yang dipakai
- **Ya** = memang harus menyentuh area itu.
- **Tidak** = secara bisnis tidak seharusnya menyentuh area itu.
- **Parsial** = ada sebagian wiring, tapi belum penuh / masih perlu audit lebih lanjut.

---

## 2. Matriks efek bisnis

| Domain | Stok | Omset | Profit / Margin | Transaction report | Daily profit report | Catatan bisnis inti |
|---|---|---:|---:|---:|---:|---|
| `purchase` | Ya | Tidak | Tidak | Ya | Tidak | Pembelian menambah stok dan masuk pembukuan pembelian, bukan penjualan. |
| `sale` | Ya | Ya | Ya | Ya | Ya | Transaksi penjualan utama. Menjadi acuan omset dan profit harian. |
| `buy_return` | Ya | Tidak | Tidak | Ya | Tidak | Retur pembelian ke supplier, diperlakukan sebagai koreksi stok + transaksi pembelian. |
| `sale_return` | Ya | Ya (koreksi negatif) | Ya (koreksi negatif) | Ya | Ya | Retur penjualan dari customer, membalik sebagian efek `sale`. |
| `duplicate_receipt` / `copy receipt` | Ya | Ya | Ya | Ya | Ya | Diperlakukan sebagai varian transaksi penjualan aktif, bukan sekadar utilitas cetak ulang nota. |
| `first_stock` | Ya | Tidak | Tidak | Ya | Tidak | Initial stock / penetapan stok awal. Bukan event penjualan. |
| `opname` | Ya | Tidak | Tidak | Ya | Tidak | Penyesuaian stok hasil audit fisik. Bukan event penjualan. |
| `expense` | Tidak | Tidak | Tidak | Ya | Tidak | Pengeluaran operasional, relevan ke pembukuan tapi bukan omset/profit sales. |
| `another_income` | Tidak | Tidak | Tidak | Ya | Tidak | Pendapatan non-penjualan, masuk transaction report tapi tidak dihitung sebagai omset penjualan harian. |

---

## 3. Keputusan penting yang harus dipegang

### A. `buy_return`
**Keputusan resmi:**
- mempengaruhi **stok**
- masuk **transaction report / neraca**
- **tidak** mempengaruhi omset penjualan
- **tidak** mempengaruhi profit / margin penjualan

Alasan:
- ini adalah retur ke supplier
- barang keluar dari apotek, jadi stok turun
- tetapi tidak terjadi event penjualan ke customer

### B. `sale_return`
**Keputusan resmi:**
- mempengaruhi **stok**
- mengoreksi **omset**
- mengoreksi **profit / margin**
- masuk **transaction report**
- masuk **daily profit report**

Alasan:
- ini adalah pembalikan sebagian / seluruh transaksi penjualan
- barang kembali ke stok
- penjualan yang tadinya tercatat harus dikurangi kembali

### C. `duplicate_receipt`
**Keputusan resmi:**
- mempengaruhi **stok**
- mempengaruhi **omset**
- mempengaruhi **profit / margin**
- masuk **transaction report**
- masuk **daily profit report**

Alasan:
- di implementasi repo ini, domain ini diperlakukan sebagai varian penjualan aktif
- bukan sekadar fitur visual untuk mencetak ulang nota lama

---

## 4. Implikasi refactor ke depan

Saat menambah fitur, merapikan handler, atau memindah logic ke service, gunakan aturan berikut:

1. **Jangan samakan semua domain hanya karena sama-sama “transaksi”.**
   - `buy_return` ≠ `sale_return`
   - `duplicate_receipt` ≈ `sale` dari sisi efek bisnis

2. **Jangan masukkan transaksi non-penjualan ke `daily_profit_reports` tanpa alasan bisnis yang jelas.**
   - `buy_return`, `expense`, `another_income`, `first_stock`, `opname` tidak boleh dipaksa mengikuti pola `sale`

3. **Setiap domain yang mempengaruhi profit harus punya dasar historis yang benar.**
   - contoh: `sale_return` memakai fondasi `hpp_snapshot` dari `sale_items`
   - hindari menghitung margin historis dari harga beli produk saat ini jika sumber historis tersedia / perlu disimpan

4. **Jika ada perubahan report, cek 2 level sekaligus:**
   - `transaction_reports`
   - `daily_profit_reports`

---

## 5. Status implementasi penting saat ini

### Sudah sehat
- `sale` → stok + omset + profit
- `sale_return` → stok + koreksi omset/profit
- `duplicate_receipt` → stok + omset + profit
- `buy_return` → stok + transaction report saja

### Harus tetap dijaga
- jangan sampai `buy_return` ikut mengotori definisi omset penjualan
- jangan sampai `sale_return` hanya mengembalikan stok tanpa mengoreksi profit report
- jangan menganggap `duplicate_receipt` sekadar alias UI bila secara bisnis ia memang transaksi penjualan

---

## 6. Cara pakai dokumen ini

Gunakan dokumen ini saat:
- audit domain transaksi baru
- menilai apakah sebuah bug ada di report atau justru di definisi bisnis
- memutuskan apakah sebuah transaksi perlu masuk `daily_profit_reports`
- menulis regression test yang menyentuh stok / omset / profit

Jika implementasi runtime bertentangan dengan dokumen ini, yang harus dilakukan adalah:
1. cek ulang source
2. cek ulang runtime dengan data uji valid
3. putuskan apakah yang salah adalah kode, report, atau asumsi bisnis
