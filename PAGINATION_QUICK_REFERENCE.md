# Ringkasan Helper PaginateWithSearchAndMonth

## Deskripsi Singkat
Helper function untuk menangani paging dengan parameter search dan month secara otomatis.

## Penggunaan Cepat

```go
// 1. Siapkan query dasar
var data []models.YourModel
query := configs.DB.Table("table_name").Select("...").Where("...")

// 2. Panggil helper
result, search, total, page, totalPages, err := helpers.PaginateWithSearchAndMonth(
    c,
    query,
    &data,
    []string{"A.purchase_id"}, // contoh: []string{"A.purchase_id", "A.payment"}
    "column_untuk_date_filter", // contoh: "A.return_date"
    1,    // defaultPage
    10,   // defaultLimit (data per halaman)
)

// 3. Handle error dan return response
if err != nil {
    return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid parameter", err.Error())
}

// 4. Format dan kirim response
return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Success", search, total, page, totalPages, 10, result)
```

## Parameter yang Diterima (Query String)
- `page=2` - Halaman yang diinginkan (default: 1)
- `search=keyword` - Keyword untuk pencarian
- `month=2025-02` - Filter bulan (format YYYY-MM, default: bulan saat ini)

## Contoh URL
```
GET /api/buy-returns?page=1&search=BUY001&month=2025-02
GET /api/buy-returns?search=ABC
GET /api/buy-returns?month=2025-01
GET /api/buy-returns  (default: halaman 1, bulan saat ini)
```

## Return Value
```go
data          interface{} // Hasil query (perlu di-cast ke tipe yang benar)
search        string      // Keyword search yang digunakan
total         int         // Total data yang sesuai filter
page          int         // Halaman saat ini
totalPages    int         // Total halaman
error         error       // Error jika ada
```

## Hal Penting
1. ✅ Gunakan **pointer** saat memanggil helper: `&data` bukan `data`
2. ✅ **Cast data** setelah menerima result: `data.([]models.YourModel)`
3. ✅ Query harus sudah **Select** dan **Where** sebelum dipassing ke helper
4. ✅ Pastikan **searchColumns** (slice) dan **dateColumn** sesuai dengan query yang dikonfigurasi
5. ✅ **Bulan default** adalah bulan saat ini jika tidak disediakan

## Perbandingan Kode

### Sebelum Helper (~30 baris)
```go
pageParam := c.Query("page")
page := 1
if p, err := strconv.Atoi(pageParam); err == nil && p > 0 { page = p }

search := strings.TrimSpace(c.Query("search"))
search = strings.ToLower(search)
query = query.Where("LOWER(A.purchase_id) ILIKE ?", "%"+search+"%")

month := strings.TrimSpace(c.Query("month"))
if month == "" { month = time.Now().Format("2006-01") }
parsedMonth, _ := time.Parse("2006-01", month)
endDate := parsedMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)
query = query.Where("A.return_date BETWEEN ? AND ?", parsedMonth, endDate)

var total int64
query.Count(&total)
query.Offset((page-1)*10).Limit(10).Scan(&data)
totalPages := int(math.Ceil(float64(total) / 10))
```

### Dengan Helper (~1 baris)
```go
data, search, total, page, totalPages, err := helpers.PaginateWithSearchAndMonth(c, query, &data, []string{"A.purchase_id"}, "A.return_date", 1, 10)
```

## File yang Telah Ditambahkan

1. **`helpers/paging_helper.go`** - Function `PaginateWithSearchAndMonth()` ditambahkan
2. **`PAGINATION_HELPER_GUIDE.md`** - Dokumentasi lengkap
3. **`controllers/pagination_example_controller.go`** - Contoh implementasi

Gunakan file `pagination_example_controller.go` sebagai referensi untuk implementasi di controller lain.
