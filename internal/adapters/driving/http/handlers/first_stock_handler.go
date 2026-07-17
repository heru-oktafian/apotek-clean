package handlers

import (
	errors "errors"
	fmt "fmt"
	http "net/http"
	strings "strings"
	time "time"

	configs "apotek-clean/configs"
	helpers "apotek-clean/helpers"
	models "apotek-clean/internal/core/entities"
	services "apotek-clean/services"
	fiber "github.com/gofiber/fiber/v2"
	gorm "gorm.io/gorm"
)

func rollbackFirstStockWithJSON(c *fiber.Ctx, tx *gorm.DB, status int, message string, err error) error {
	tx.Rollback()
	return helpers.JSONResponse(c, status, message, err)
}

func ensureFirstStockBranchExists(tx *gorm.DB, branchID string) error {
	var branch models.Branch
	return tx.Where("id = ?", branchID).First(&branch).Error
}

func applyFirstStockQuotaIfNeeded(tx *gorm.DB, subscriptionType, branchID string) error {
	if subscriptionType != "quota" {
		return nil
	}
	return ensureFirstStockBranchExists(tx, branchID)
}

func finalizeFirstStockTransaction(tx *gorm.DB, firstStock models.FirstStocks, subscriptionType string) error {
	if err := applyFirstStockQuotaIfNeeded(tx, subscriptionType, firstStock.BranchID); err != nil {
		return err
	}
	return tx.Commit().Error
}

// CreateFirstStock Function
func CreateFirstStock(c *fiber.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	db := configs.DB

	// Ambil informasi dari token
	branchID, _ := services.GetBranchID(c)
	userID, _ := services.GetUserID(c)
	generatedID := helpers.GenerateID("FST")

	// Ambil input dari body
	var input models.FirstStockInput
	if err := c.BodyParser(&input); err != nil {
		return helpers.JSONResponse(c, http.StatusBadRequest, "Input first stock tidak valid", err)
	}

	parsedDate, err := services.ParseFirstStockDate(input.FirstStockDate, nowWIB)
	if err != nil {
		return helpers.JSONResponse(c, http.StatusBadRequest, "Format tanggal tidak valid. Gunakan YYYY-MM-DD", err.Error())
	}

	// Map ke struct model
	first_stock := models.FirstStocks{
		ID:              generatedID,
		Description:     input.Description,
		BranchID:        branchID,
		UserID:          userID,
		FirstStockDate:  parsedDate,
		TotalFirstStock: 0,
		CreatedAt:       nowWIB,
		UpdatedAt:       nowWIB,
	}

	// Simpan first_stock
	if err := db.Create(&first_stock).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal membuat first stock", err)
	}

	// Buat laporan
	if err := SyncFirstStockReport(db, first_stock); err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal menyinkronkan laporan first stock", err)
	}

	return helpers.JSONResponse(c, http.StatusOK, "First stock berhasil dibuat", first_stock)
}

// UpdateFirstStock Function
func UpdateFirstStock(c *fiber.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	db := configs.DB
	id := c.Params("id")

	// Cari data first_stock lama
	var first_stock models.FirstStocks
	if err := db.First(&first_stock, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusNotFound, "First stock tidak ditemukan", err)
	}

	// Gunakan struct input
	var input models.FirstStockInput
	if err := c.BodyParser(&input); err != nil {
		return helpers.JSONResponse(c, http.StatusBadRequest, "Input first stock tidak valid", err)
	}

	if input.FirstStockDate != "" {
		parsedDate, err := services.ParseFirstStockDate(input.FirstStockDate, nowWIB)
		if err != nil {
			return helpers.JSONResponse(c, http.StatusBadRequest, "Format tanggal tidak valid. Gunakan YYYY-MM-DD", err)
		}
		first_stock.FirstStockDate = parsedDate
	}

	// Cek dan update Payment
	if input.Payment != "" {
		first_stock.Payment = models.PaymentStatus(input.Payment)
	}

	first_stock.UpdatedAt = nowWIB

	// Hitung ulang total dari first_stock items
	var items []models.FirstStockItems
	if err := db.Where("first_stock_id = ?", id).Find(&items).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal mengambil item first stock", err)
	}

	first_stock.TotalFirstStock = services.SumFirstStockItems(items)

	// Cek dan update Description
	if input.Description != "" {
		first_stock.Description = input.Description
	}

	// Simpan perubahan
	if err := db.Save(&first_stock).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal memperbarui first stock", err)
	}

	// Sync report
	if err := SyncFirstStockReport(db, first_stock); err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal menyinkronkan laporan first stock", err)
	}

	return helpers.JSONResponse(c, http.StatusOK, "First stock berhasil diperbarui", first_stock)
}

// DeleteFirstStock Function
func DeleteFirstStock(c *fiber.Ctx) error {
	db := configs.DB
	id := c.Params("id")

	// Ambil first_stock
	var first_stock models.FirstStocks
	if err := db.First(&first_stock, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusNotFound, "First stock tidak ditemukan", err)
	}

	// Ambil item-item dan rollback stok
	var items []models.FirstStockItems
	if err := db.Where("first_stock_id = ?", id).Find(&items).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal mengambil item first stock", err)
	}

	for _, item := range items {
		// Kurangi stok ke produk
		if err := services.ReduceProductStock(db, item.ProductId, item.Qty); err != nil {
			return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal mengurangi stok produk", err)
		}
	}

	// Hapus semua item dari pembelian
	if err := db.Where("first_stock_id = ?", id).Delete(&models.FirstStockItems{}).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal menghapus item first stock", err)
	}

	// Hapus laporan transaksi terkait
	if err := db.Where("id = ? AND transaction_type = ?", first_stock.ID, models.FirstStock).Delete(&models.TransactionReports{}).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal menghapus laporan transaksi", err)
	}

	// Hapus first_stock
	if err := db.Delete(&first_stock).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal menghapus first stock", err)
	}

	return helpers.JSONResponse(c, http.StatusOK, "First stock berhasil dihapus", first_stock)
}

// CreateFirstStockItem Function
func CreateFirstStockItem(c *fiber.Ctx) error {
	db := configs.DB
	var item models.FirstStockItems

	if err := c.BodyParser(&item); err != nil {
		return helpers.JSONResponse(c, http.StatusBadRequest, "Input first stock tidak valid", err)
	}

	// Cek apakah item dengan first_stock_id dan product_id sudah ada
	var existing models.FirstStockItems
	err := db.Where("first_stock_id = ? AND product_id = ?", item.FirstStockId, item.ProductId).First(&existing).Error
	if err == nil {
		// Sudah ada: update qty dan sub_total
		existing.Qty += item.Qty
		existing.SubTotal = existing.Qty * existing.Price // asumsi pakai harga awal

		if err := db.Save(&existing).Error; err != nil {
			return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal memperbarui item first stock", err)
		}

		// Tambah stok
		if err := services.AddProductStock(db, item.ProductId, item.Qty); err != nil {
			return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal menambahkan stok produk", err)
		}

		// Update harga produk jika harga baru lebih tinggi dari yang tersimpan di tabel products
		if err := services.UpdateProductPriceIfHigher(db, item.ProductId, item.Price); err != nil {
			return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal memperbarui harga produk", err)
		}

		go func() {
			if err := RecalculateTotalFirstStock(db, item.FirstStockId); err != nil {
				fmt.Printf("Failed to recalculate total FirstStock asynchronously: %v\n", err)
			}
		}()

		return helpers.JSONResponse(c, http.StatusOK, "Item first stock berhasil diperbarui", existing)

	} else if err != gorm.ErrRecordNotFound {
		// Error selain record not found
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal mengambil item first stock", err)
	}

	// Data belum ada, buat item baru
	if item.ID == "" {
		item.ID = helpers.GenerateID("FSI")
	}
	item.SubTotal = item.Qty * item.Price

	if err := db.Create(&item).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal membuat item first stock", err)
	}

	if err := services.AddProductStock(db, item.ProductId, item.Qty); err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal menambahkan stok produk", err)
	}

	if err := services.UpdateProductPriceIfHigher(db, item.ProductId, item.Price); err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal memperbarui harga produk", err)
	}

	go func() {
		if err := RecalculateTotalFirstStock(db, item.FirstStockId); err != nil {
			fmt.Printf("Failed to recalculate total FirstStock asynchronously: %v\n", err)
		}
	}()

	return helpers.JSONResponse(c, http.StatusOK, "Item first stock berhasil ditambahkan", item)
}

// Update FirstStockItem
func UpdateFirstStockItem(c *fiber.Ctx) error {
	db := configs.DB
	id := c.Params("id")

	var existingItem models.FirstStockItems
	if err := db.First(&existingItem, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusNotFound, "Item tidak ditemukan", err)
	}

	var updatedItem models.FirstStockItems
	if err := c.BodyParser(&updatedItem); err != nil {
		return helpers.JSONResponse(c, http.StatusBadRequest, "Input first stock tidak valid", err)
	}

	// Rollback stok lama
	if err := services.ReduceProductStock(db, existingItem.ProductId, existingItem.Qty); err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal rollback stok produk", err)
	}

	// Tambah stok baru
	if err := services.AddProductStock(db, updatedItem.ProductId, updatedItem.Qty); err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal menambahkan stok produk", err)
	}

	// Update item
	existingItem.ProductId = updatedItem.ProductId
	existingItem.Qty = updatedItem.Qty
	existingItem.Price = updatedItem.Price
	existingItem.SubTotal = updatedItem.Price * updatedItem.Qty

	if err := db.Save(&existingItem).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal memperbarui item first stock", err)
	}

	// Update harga produk jika harga item lebih tinggi
	if err := services.UpdateProductPriceIfHigher(db, updatedItem.ProductId, updatedItem.Price); err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal memperbarui harga produk", err)
	}

	go func() {
		if err := RecalculateTotalFirstStock(db, existingItem.FirstStockId); err != nil {
			fmt.Printf("Failed to recalculate total FirstStock asynchronously: %v\n", err)
		}
	}()

	return helpers.JSONResponse(c, http.StatusOK, "Item first stock berhasil diperbarui", existingItem)
}

// Delete FirstStockItem
func DeleteFirstStockItem(c *fiber.Ctx) error {
	db := configs.DB
	id := c.Params("id")

	var item models.FirstStockItems
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusNotFound, "Item tidak ditemukan", err)
	}

	// Subtract stok
	if err := services.ReduceProductStock(db, item.ProductId, item.Qty); err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal rollback stok produk", err)
	}

	// Hapus item
	if err := db.Delete(&item).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to delete FirstStock item", err)
	}

	go func() {
		if err := RecalculateTotalFirstStock(db, item.FirstStockId); err != nil {
			fmt.Printf("Failed to recalculate total FirstStock asynchronously: %v\n", err)
		}
	}()

	return helpers.JSONResponse(c, http.StatusOK, "Item first stock berhasil dihapus", item)
}

// Get All FirstStocks tampilkan semua first_stock
func GetAllFirstStocks(c *fiber.Ctx) error {
	// Dapatkan waktu sekarang di WIB
	nowWIB := time.Now().In(configs.Location)

	// Get branch id
	branch_id, _ := services.GetBranchID(c)

	month := strings.TrimSpace(c.Query("month"))

	// Jika month kosong, isi dengan bulan ini (format YYYY-MM)
	if month == "" {
		month = nowWIB.Format("2006-01")
	}

	startDate, err := time.Parse("2006-01", month)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Format bulan tidak valid. Gunakan YYYY-MM", nil)
	}
	endDate := startDate.AddDate(0, 1, 0)

	var FirstStocks []models.AllFirstStocks
	var total int

	// Query dasar
	query := configs.DB.Table("first_stocks pur").
		Select("pur.id, pur.description, pur.first_stock_date, pur.total_first_stock, pur.payment").
		Where("pur.branch_id = ?", branch_id).
		Where("pur.first_stock_date >= ? AND pur.first_stock_date < ?", startDate, endDate).
		Order("pur.created_at DESC")

	_, search, total, page, totalPages, limit, err := helpers.Paginate(c, query, &FirstStocks, []string{"pur.description ILIKE ?"})
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data first stock", err.Error())
	}

	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Data first stock berhasil diambil", search, int(total), page, int(totalPages), int(limit), FirstStocks)
}

// GetAllFirstStockItems tampilkan semua item berdasarkan first_stock_id tanpa pagination
func GetAllFirstStockItems(c *fiber.Ctx) error {
	// Get FirstStock id dari param
	first_stockID := c.Params("id")

	search := strings.TrimSpace(c.Query("search"))

	// Jika ada search key, tambahkan filter WHERE
	if search != "" {
		search = strings.ToLower(search)
	}

	var FirstStockItems []models.AllFirstStockItems

	// Query dasar
	query := configs.DB.Table("first_stock_items pit").
		Select("pit.id, pit.first_stock_id, pit.product_id, pro.name AS product_name, pit.price, pit.qty, un.name AS unit_name, pit.sub_total").
		Joins("LEFT JOIN products pro ON pro.id = pit.product_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("pit.first_stock_id = ?", first_stockID).
		Order("pro.name ASC")

	// Jika ada search key, tambahkan filter WHERE
	if search != "" {
		search = strings.ToLower(search)
		query = query.Where("LOWER(pro.name) LIKE ?", "%"+search+"%")
	}

	// Eksekusi query
	if err := query.Scan(&FirstStockItems).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal mengambil item first stock", err)
	}

	return helpers.JSONResponse(c, http.StatusOK, "Data item first stock berhasil diambil", FirstStockItems)
}

// GetFirstStockWithItems menampilkan satu first_stock beserta semua item-nya
func GetFirstStockWithItems(c *fiber.Ctx) error {
	db := configs.DB

	// Ambil ID pembelian dari parameter URL
	first_stockID := c.Params("id")

	// Struct untuk data utama first_stock
	var first_stock models.AllFirstStocks

	// Ambil data first_stock
	err := db.Table("first_stocks pur").
		Select("pur.id, pur.description, pur.first_stock_date, pur.total_first_stock, pur.payment").
		Where("pur.id = ?", first_stockID).
		Scan(&first_stock).Error

	if err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal mengambil data first stock", err)
	}

	// Ambil item pembelian terkait
	var items []models.AllFirstStockItems
	err = db.Table("first_stock_items pit").
		Select("pit.id, pit.first_stock_id, pit.product_id, pro.name AS product_name, pit.price, pit.qty, un.name AS unit_name, pit.sub_total").
		Joins("LEFT JOIN products pro ON pro.id = pit.product_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("pit.first_stock_id = ?", first_stockID).
		Order("pro.name ASC").
		Scan(&items).Error

	if err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal mengambil item first stock", err)
	}

	// Format tanggal pembelian ke dd-mm-yyyy
	formattedDate := first_stock.FirstStockDate.Format("02-01-2006")

	return JSONFirstStockWithItemsResponse(c, http.StatusOK, "Data first stock berhasil diambil", first_stockID, first_stock.Description, formattedDate, first_stock.TotalFirstStock, string(first_stock.Payment), items)
}

// CreateFirstStockTransaction controller
func CreateFirstStockTransaction(c *fiber.Ctx) error {
	nowWIB := time.Now().In(configs.Location)

	subscriptionType, _ := services.GetClaimsToken(c, "subscription_type")
	branchID, _ := services.GetBranchID(c)
	userID, _ := services.GetUserID(c)

	db := configs.DB
	var req models.FirstStockTransactionRequest
	err := c.BodyParser(&req)
	if err != nil {
		return helpers.JSONResponse(c, http.StatusBadRequest, "Body permintaan tidak valid", err)
	}

	// Set Payment secara default karena ini 'first_stock' (tidak ada pembiayaan)
	// Anda bisa pilih "nocost" atau jika punya models.NoCost, gunakan itu.
	var paymentStatus models.PaymentStatus = "nocost" // Default ke nocost

	// Inisialisasi header FirstStock dengan data dari token dan default payment
	firstStockHeader := models.FirstStocks{
		UserID:   userID,
		BranchID: branchID,
		Payment:  paymentStatus,
	}

	// --- VALIDASI INPUT ---
	// Validasi input header dan item
	if err = helpers.ValidateStruct(req.FirstStock); err != nil {
		return helpers.JSONResponse(c, http.StatusBadRequest, "Validasi header first stock gagal", err)
	}
	for _, item := range req.FirstStockItems {
		if err = helpers.ValidateStruct(item); err != nil {
			return helpers.JSONResponse(c, http.StatusBadRequest, "Validasi satu atau lebih item first stock gagal", err)
		}
		// Validasi manual tambahan karena tag required di-relax
		if item.ProductId == "" || item.UnitId == "" {
			return helpers.JSONResponse(c, http.StatusBadRequest, "product_id dan unit_id wajib diisi untuk semua item", nil)
		}
		if item.Qty <= 0 {
			return helpers.JSONResponse(c, http.StatusBadRequest, "qty harus lebih besar dari 0", nil)
		}
	}
	// --- AKHIR VALIDASI INPUT ---

	// Parse FirstStockDate
	// Parse FirstStockDate
	var parsedFirstStockDate time.Time
	if req.FirstStock.FirstStockDate == "" {
		parsedFirstStockDate = nowWIB
	} else {
		parsedFirstStockDate, err = time.Parse("2006-01-02", req.FirstStock.FirstStockDate)
		if err != nil {
			return helpers.JSONResponse(c, http.StatusBadRequest, "Format first_stock_date tidak valid. Gunakan YYYY-MM-DD.", err.Error())
		}
	}

	// Mengisi detail FirstStocks dari request dan data token/default
	firstStockHeader.ID = helpers.GenerateID("FST") // Generate ID untuk First Stock
	firstStockHeader.Description = req.FirstStock.Description
	firstStockHeader.FirstStockDate = parsedFirstStockDate
	firstStockHeader.CreatedAt = nowWIB
	firstStockHeader.UpdatedAt = nowWIB

	// --- Proses Penyimpanan Data (Dalam Transaksi Database) ---
	tx := db.Begin()
	if tx.Error != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal memulai transaksi database", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var calculatedTotalFirstStock int
	var firstStockItemsToCreate []models.FirstStockItems
	var firstStockItemsForResponse []models.FirstStockItemResponse // Slice untuk data respons

	// var stockTracksToCreate []models.StockTracks

	for _, reqItem := range req.FirstStockItems {
		parsedExpiredDate, err := time.Parse("2006-01-02", reqItem.ExpiredDate)
		if err != nil {
			return rollbackFirstStockWithJSON(c, tx, http.StatusBadRequest, fmt.Sprintf("Format expired_date tidak valid untuk produk %s. Gunakan YYYY-MM-DD.", reqItem.ProductId), err)
		}

		lookup, err := services.LookupFirstStockDependencies(tx, firstStockHeader.BranchID, reqItem.ProductId, reqItem.UnitId)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				if tx.Where("id = ? AND branch_id = ?", reqItem.ProductId, firstStockHeader.BranchID).First(&models.Product{}).Error == nil {
					return rollbackFirstStockWithJSON(c, tx, http.StatusNotFound, fmt.Sprintf("Unit with ID %s not found", reqItem.UnitId), err)
				}
				return rollbackFirstStockWithJSON(c, tx, http.StatusNotFound, fmt.Sprintf("Product with ID %s not found in branch %s", reqItem.ProductId, firstStockHeader.BranchID), err)
			}
			return rollbackFirstStockWithJSON(c, tx, http.StatusInternalServerError, "Gagal mengambil dependensi item first stock", err)
		}

		preparedItem := services.PrepareFirstStockItem(helpers.GenerateID("FSI"), firstStockHeader.ID, reqItem, lookup, parsedExpiredDate)
		firstStockItemsToCreate = append(firstStockItemsToCreate, preparedItem.Item)
		firstStockItemsForResponse = append(firstStockItemsForResponse, preparedItem.Response)

		err = tx.Model(&models.Product{}).Where("id = ?", lookup.Product.ID).Updates(preparedItem.ProductUpdate).Error
		if err != nil {
			return rollbackFirstStockWithJSON(c, tx, http.StatusInternalServerError, fmt.Sprintf("Gagal memperbarui detail produk (stok/expired_date) untuk produk %s", lookup.Product.Name), err)
		}

		calculatedTotalFirstStock += preparedItem.SubTotal
	}

	firstStockHeader.TotalFirstStock = calculatedTotalFirstStock

	// Simpan data FirstStocks
	err = tx.Create(&firstStockHeader).Error
	if err != nil {
		return rollbackFirstStockWithJSON(c, tx, http.StatusInternalServerError, "Gagal membuat data first stock", err)
	}

	// Simpan FirstStockItems dalam batch
	err = tx.CreateInBatches(&firstStockItemsToCreate, len(firstStockItemsToCreate)).Error
	if err != nil {
		return rollbackFirstStockWithJSON(c, tx, http.StatusInternalServerError, "Gagal membuat item first stock", err)
	}

	// PENTING: TransactionReports dan DailyProfitReport TIDAK relevan untuk First Stock
	// Karena ini bukan transaksi finansial atau penjualan/pembelian berbiaya,
	// bagian untuk membuat TransactionReports atau mengupdate DailyProfitReport dihapus.

	err = finalizeFirstStockTransaction(tx, firstStockHeader, subscriptionType)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return rollbackFirstStockWithJSON(c, tx, http.StatusNotFound, fmt.Sprintf("Branch with ID %s not found", firstStockHeader.BranchID), err)
		}
		return rollbackFirstStockWithJSON(c, tx, http.StatusInternalServerError, "Gagal menyelesaikan transaksi first stock", err)
	}

	// --- Mengkonstruksi Objek Respon ---
	response := models.FirstStockTransactionResponse{
		FirstStock: models.FirstStockOutput{
			ID:              firstStockHeader.ID,
			Description:     firstStockHeader.Description,
			FirstStockDate:  firstStockHeader.FirstStockDate.Format("2006-01-02"), // Format YYYY-MM-DD
			BranchID:        firstStockHeader.BranchID,
			TotalFirstStock: firstStockHeader.TotalFirstStock,
			Payment:         string(firstStockHeader.Payment),
			UserID:          firstStockHeader.UserID,
			CreatedAt:       firstStockHeader.CreatedAt.Format("2006-01-02"), // Format YYYY-MM-DD
			UpdatedAt:       firstStockHeader.UpdatedAt.Format("2006-01-02"), // Format YYYY-MM-DD
		},
		FirstStockItems: firstStockItemsForResponse,
	}
	// --- Akhir Mengkonstruksi Objek Respon ---

	return helpers.JSONResponse(c, http.StatusCreated, "Transaksi first stock berhasil dibuat", response)
}

// Insert atau update laporan transaksi berdasarkan FirstStocks / Pengeluaran
func SyncFirstStockReport(db *gorm.DB, first_stock models.FirstStocks) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	// Siapkan data report dari FirstStock
	report := models.TransactionReports{
		ID:              first_stock.ID,
		TransactionType: models.FirstStock,
		UserID:          first_stock.UserID,
		BranchID:        first_stock.BranchID,
		Total:           first_stock.TotalFirstStock,
		CreatedAt:       first_stock.CreatedAt,
		UpdatedAt:       first_stock.UpdatedAt,
		Payment:         first_stock.Payment,
	}

	var existing models.TransactionReports
	err := db.Take(&existing, "id = ?", report.ID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Insert
		return db.Create(&report).Error
	}
	if err != nil {
		return err
	}

	// Jika ditemukan, lakukan update pada kolom yang dibutuhkan
	existing.Total = report.Total
	existing.UpdatedAt = nowWIB
	existing.Payment = report.Payment

	return db.Save(&existing).Error
}

func RecalculateTotalFirstStock(db *gorm.DB, first_stockID string) error {
	var total int64

	// Hitung total sub_total dari first_stock_items
	err := db.Model(&models.FirstStockItems{}).
		Where("first_stock_id = ?", first_stockID).
		Select("COALESCE(SUM(sub_total), 0)").
		Scan(&total).Error

	if err != nil {
		return err
	}

	// Update ke first_stocks
	if err := db.Model(&models.FirstStocks{}).
		Where("id = ?", first_stockID).
		Update("total_first_stock", total).Error; err != nil {
		return err
	}

	// Ambil first_stock lengkap buat update report
	var first_stock models.FirstStocks
	if err := db.First(&first_stock, "id = ?", first_stockID).Error; err != nil {
		return err
	}

	// Update transaction_reports juga
	if err := SyncFirstStockReport(db, first_stock); err != nil {
		return err
	}

	return nil
}

// JSONFirstStockWithItemsResponse sends a standard JSON response format / structure
func JSONFirstStockWithItemsResponse(c *fiber.Ctx, status int, message string, first_stock_id string, description string, first_stock_date string, total_first_stock int, payment string, items interface{}) error {
	resp := models.ResponseFirstStockWithItemsResponse{
		Status:          http.StatusText(status),
		Message:         message,
		FirstStockId:    first_stock_id,
		Description:     description,
		FirstStockDate:  first_stock_date,
		TotalFirstStock: total_first_stock,
		Payment:         payment,
		Items:           items,
	}
	return helpers.JSONResponse(c, status, message, resp)
}
