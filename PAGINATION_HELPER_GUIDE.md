# Panduan Penggunaan PaginateWithSearchAndMonth Helper

## Deskripsi
Function `PaginateWithSearchAndMonth` adalah helper untuk menangani pagination dengan mendukung parameter `search` dan `month` secara bersamaan. Helper ini dirancang untuk menyederhanakan kode controller ketika menangani paging dengan filter bulan dan pencarian.

## Signature Function
```go
func PaginateWithSearchAndMonth(
	c *fiber.Ctx,
	query *gorm.DB,
	model interface{},
	searchColumns []string,
	dateColumn string,
	defaultPage int,
	defaultLimit int,
) (interface{}, string, int, int, int, error)
```

## Parameter
- **c**: Fiber context dari HTTP request
- **query**: GORM query builder yang sudah dikonfigurasi (dengan select, where, order, dll)
- **model**: Pointer ke slice struct untuk menampung hasil data (contoh: `&[]models.AllBuyReturns{}`)
- **searchColumns**: Slice of string untuk kolom filter search (contoh: `[]string{"A.purchase_id", "A.payment"}`)
- **dateColumn**: Nama kolom tanggal untuk filter bulan (contoh: `"A.return_date"`)
- **defaultPage**: Halaman default jika parameter page tidak disediakan (biasanya `1`)
- **defaultLimit**: Jumlah data per halaman (contoh: `10`)

## Return Value
1. **data** (interface{}): Hasil query yang sudah dipaginate
2. **search** (string): Keyword search yang digunakan
3. **total** (int): Total data yang sesuai dengan filter
4. **currentPage** (int): Halaman saat ini
5. **totalPages** (int): Total halaman berdasarkan hasil filter
6. **error** (error): Error jika ada

## Query Parameter yang Didukung
- **page**: Nomor halaman (default: 1)
- **search**: Keyword untuk filter pencarian
- **month**: Format bulan YYYY-MM (default: bulan saat ini)

## Contoh Penggunaan

### Sebelum (Kode Awal)
```go
func GetAllBuyReturns(c *fiber.Ctx) error {
	nowWIB := time.Now().In(configs.Location)
	branchID, _ := services.GetBranchID(c)

	pageParam := c.Query("page")
	search := strings.TrimSpace(c.Query("search"))

	page := 1
	if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
		page = p
	}

	limit := 10
	offset := (page - 1) * limit

	month := strings.TrimSpace(c.Query("month"))
	if month == "" {
		month = nowWIB.Format("2006-01")
	}

	var buyReturnsFromDB []models.AllBuyReturns
	var total int64

	query := configs.DB.Table("buy_returns A").
		Select("A.id, A.purchase_id, A.return_date, A.payment, A.total_return").
		Where("A.branch_id = ? ", branchID).
		Order("A.created_at DESC")

	if search != "" {
		search = strings.ToLower(search)
		query = query.Where("LOWER(A.purchase_id) ILIKE ?", "%"+search+"%")
	}

	if month != "" {
		parsedMonth, err := time.Parse("2006-01", month)
		if err != nil {
			return helpers.JSONResponse(c, fiber.StatusBadRequest, "Format bulan tidak valid", "...")
		}
		startDate := parsedMonth
		endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
		query = query.Where("A.return_date BETWEEN ? AND ?", startDate, endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil retur pembelian", "...")
	}

	if err := query.Offset(offset).Limit(limit).Scan(&buyReturnsFromDB).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil retur pembelian", "...")
	}

	var formattedBuyReturnsData []models.BuyReturnsResponse
	for _, buyReturn := range buyReturnsFromDB {
		formattedBuyReturnsData = append(formattedBuyReturnsData, models.BuyReturnsResponse{
			ID:          buyReturn.ID,
			PurchaseId:  buyReturn.PurchaseId,
			ReturnDate:  helpers.FormatIndonesianDate(buyReturn.ReturnDate),
			TotalReturn: buyReturn.TotalReturn,
			Payment:     string(buyReturn.Payment),
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return helpers.JSONResponseGetAll(
		c,
		fiber.StatusOK,
		"Data retur pembelian berhasil diambil",
		search,
		int(total),
		page,
		totalPages,
		limit,
		formattedBuyReturnsData,
	)
}
```

### Sesudah (Dengan Helper)
```go
func GetAllBuyReturns(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)

	var buyReturnsFromDB []models.AllBuyReturns

	query := configs.DB.Table("buy_returns A").
		Select("A.id, A.purchase_id, A.return_date, A.payment, A.total_return").
		Where("A.branch_id = ? ", branchID).
		Order("A.created_at DESC")

	// Gunakan helper untuk pagination, search, dan month
	data, search, total, page, totalPages, err := helpers.PaginateWithSearchAndMonth(
		c,
		query,
		&buyReturnsFromDB,
		[]string{"A.purchase_id"}, // searchColumns
		"A.return_date",        // dateColumn
		1,                      // defaultPage
		10,                     // defaultLimit
	)

	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Filter tidak valid", err.Error())
	}

	// Format data sebelum mengirim response
	var formattedBuyReturnsData []models.BuyReturnsResponse
	for _, buyReturn := range data.([]models.AllBuyReturns) {
		formattedBuyReturnsData = append(formattedBuyReturnsData, models.BuyReturnsResponse{
			ID:          buyReturn.ID,
			PurchaseId:  buyReturn.PurchaseId,
			ReturnDate:  helpers.FormatIndonesianDate(buyReturn.ReturnDate),
			TotalReturn: buyReturn.TotalReturn,
			Payment:     string(buyReturn.Payment),
		})
	}

	return helpers.JSONResponseGetAll(
		c,
		fiber.StatusOK,
		"Data retur pembelian berhasil diambil",
		search,
		total,
		page,
		totalPages,
		10,
		formattedBuyReturnsData,
	)
}
```

## Keuntungan Menggunakan Helper

1. **Kode Lebih Ringkas**: Mengurangi boilerplate code yang repetitif
2. **Konsistensi**: Implementasi pagination yang sama di semua controller
3. **Maintainability**: Jika ada perubahan logic pagination, hanya perlu update di satu tempat
4. **Error Handling**: Penanganan error untuk format bulan yang tidak valid sudah terintegrasi
5. **Default Value**: Otomatis menggunakan bulan saat ini jika parameter month tidak disediakan

## Contoh URL Request

```
GET /api/buy-returns?page=2&search=BUY001&month=2025-02
GET /api/buy-returns?page=1&search=
GET /api/buy-returns?month=2025-01
GET /api/buy-returns (akan menggunakan default bulan saat ini)
```

## Error Handling

Helper akan mengembalikan error jika:
1. Format bulan tidak sesuai YYYY-MM
2. Query database gagal (count atau scan)

Contoh penanganan error:
```go
if err != nil {
	return helpers.JSONResponse(c, fiber.StatusBadRequest, "Format tidak valid", err.Error())
}
```

## Tips Penggunaan

1. **Jangan lupa cast data**: Hasil dari `data` adalah interface{}, pastikan di-cast ke slice struct yang benar
   ```go
   buyReturns := data.([]models.AllBuyReturns)
   ```

2. **Query harus sudah dikonfigurasi**: Helper hanya menangani filter search dan month, konfigurasi query lain (Select, Join, Where) tetap di controller
   ```go
   query := configs.DB.Table("buy_returns A").
       Select("A.id, A.purchase_id, A.return_date, A.payment, A.total_return").
       Where("A.branch_id = ? ", branchID).
       Order("A.created_at DESC")
   ```

3. **Page harus lebih dari 0**: Helper otomatis validasi page, jika page <= 0 akan digunakan defaultPage

4. **Search case-insensitive**: Pencarian sudah otomatis convert ke lowercase untuk case-insensitive search

## Use Case Lainnya

Helper ini dapat digunakan untuk limit berbeda di controller yang berbeda:

```go
// Untuk list dengan limit 10
helpers.PaginateWithSearchAndMonth(c, query, &data, []string{"column"}, "date_column", 1, 10)

// Untuk list dengan limit 20
helpers.PaginateWithSearchAndMonth(c, query, &data, []string{"column"}, "date_column", 1, 20)

// Untuk list dengan limit 50
helpers.PaginateWithSearchAndMonth(c, query, &data, []string{"column"}, "date_column", 1, 50)
```
