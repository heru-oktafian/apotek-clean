package controllers

import (
	errors "errors"
	fmt "fmt"
	http "net/http"
	strings "strings"
	time "time"

	fiber "github.com/gofiber/fiber/v2"
	configs "github.com/heru-oktafian/fiber-apotek/configs"
	helpers "github.com/heru-oktafian/fiber-apotek/helpers"
	models "github.com/heru-oktafian/fiber-apotek/models"
	services "github.com/heru-oktafian/fiber-apotek/services"
	gorm "gorm.io/gorm"
)

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
		return helpers.JSONResponse(c, http.StatusBadRequest, "Invalid input", err)
	}

	// Parse tanggal
	layout := "2006-01-02" // format harus YYYY-MM-DD
	parsedDate, err := time.Parse(layout, input.FirstStockDate)
	if err != nil {
		return helpers.JSONResponse(c, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD", err.Error())
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
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to create FirstStock", err)
	}

	// Buat laporan
	if err := SyncFirstStockReport(db, first_stock); err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to sync FirstStock report", err)
	}

	return helpers.JSONResponse(c, http.StatusOK, "FirstStock created successfully", first_stock)
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
		return helpers.JSONResponse(c, http.StatusNotFound, "FirstStock not found", err)
	}

	// Gunakan struct input
	var input models.FirstStockInput
	if err := c.BodyParser(&input); err != nil {
		return helpers.JSONResponse(c, http.StatusBadRequest, "Invalid input", err)
	}

	// Cek dan update FirstStockDate
	if input.FirstStockDate != "" {
		layout := "2006-01-02"
		parsedDate, err := time.Parse(layout, input.FirstStockDate)
		if err != nil {
			return helpers.JSONResponse(c, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD", err)
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
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to retrieve FirstStock items", err)
	}

	if len(items) == 0 {
		first_stock.TotalFirstStock = 0
	} else {
		total := 0
		for _, item := range items {
			total += item.SubTotal
		}
		first_stock.TotalFirstStock = total
	}

	// Cek dan update Description
	if input.Description != "" {
		first_stock.Description = input.Description
	}

	// Simpan perubahan
	if err := db.Save(&first_stock).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to update FirstStock", err)
	}

	// Sync report
	if err := SyncFirstStockReport(db, first_stock); err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to sync FirstStock report", err)
	}

	return helpers.JSONResponse(c, http.StatusOK, "FirstStock updated successfully", first_stock)
}

// DeleteFirstStock Function
func DeleteFirstStock(c *fiber.Ctx) error {
	db := configs.DB
	id := c.Params("id")

	// Ambil first_stock
	var first_stock models.FirstStocks
	if err := db.First(&first_stock, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusNotFound, "FirstStock not found", err)
	}

	// Ambil item-item dan rollback stok
	var items []models.FirstStockItems
	if err := db.Where("first_stock_id = ?", id).Find(&items).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to retrieve FirstStock items", err)
	}

	for _, item := range items {
		// Kurangi stok ke produk
		if err := services.ReduceProductStock(db, item.ProductId, item.Qty); err != nil {
			return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to reduce product stock", err)
		}
	}

	// Hapus semua item dari pembelian
	if err := db.Where("first_stock_id = ?", id).Delete(&models.FirstStockItems{}).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to delete FirstStock items", err)
	}

	// Hapus laporan transaksi terkait
	if err := db.Where("id = ? AND transaction_type = ?", first_stock.ID, models.FirstStock).Delete(&models.TransactionReports{}).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to delete TransactionReports", err)
	}

	// Hapus first_stock
	if err := db.Delete(&first_stock).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to delete FirstStock", err)
	}

	return helpers.JSONResponse(c, http.StatusOK, "FirstStock deleted successfully", first_stock)
}

// CreateFirstStockItem Function
func CreateFirstStockItem(c *fiber.Ctx) error {
	db := configs.DB
	var item models.FirstStockItems

	if err := c.BodyParser(&item); err != nil {
		return helpers.JSONResponse(c, http.StatusBadRequest, "Invalid input", err)
	}

	// Cek apakah item dengan first_stock_id dan product_id sudah ada
	var existing models.FirstStockItems
	err := db.Where("first_stock_id = ? AND product_id = ?", item.FirstStockId, item.ProductId).First(&existing).Error
	if err == nil {
		// Sudah ada: update qty dan sub_total
		existing.Qty += item.Qty
		existing.SubTotal = existing.Qty * existing.Price // asumsi pakai harga awal

		if err := db.Save(&existing).Error; err != nil {
			return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to update FirstStock item", err)
		}

		// Tambah stok
		if err := services.AddProductStock(db, item.ProductId, item.Qty); err != nil {
			return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to add product stock", err)
		}

		// Update harga produk jika harga baru lebih tinggi dari yang tersimpan di tabel products
		if err := services.UpdateProductPriceIfHigher(db, item.ProductId, item.Price); err != nil {
			return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to update product price", err)
		}

		go func() {
			if err := RecalculateTotalFirstStock(db, item.FirstStockId); err != nil {
				fmt.Printf("Failed to recalculate total FirstStock asynchronously: %v\n", err)
			}
		}()

		return helpers.JSONResponse(c, http.StatusOK, "Item updated successfully", existing)

	} else if err != gorm.ErrRecordNotFound {
		// Error selain record not found
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to retrieve FirstStock item", err)
	}

	// Data belum ada, buat item baru
	if item.ID == "" {
		item.ID = helpers.GenerateID("FSI")
	}
	item.SubTotal = item.Qty * item.Price

	if err := db.Create(&item).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to create FirstStock item", err)
	}

	if err := services.AddProductStock(db, item.ProductId, item.Qty); err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to add product stock", err)
	}

	if err := services.UpdateProductPriceIfHigher(db, item.ProductId, item.Price); err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to update product price", err)
	}

	go func() {
		if err := RecalculateTotalFirstStock(db, item.FirstStockId); err != nil {
			fmt.Printf("Failed to recalculate total FirstStock asynchronously: %v\n", err)
		}
	}()

	return helpers.JSONResponse(c, http.StatusOK, "Item added successfully", item)
}

// Update FirstStockItem
func UpdateFirstStockItem(c *fiber.Ctx) error {
	db := configs.DB
	id := c.Params("id")

	var existingItem models.FirstStockItems
	if err := db.First(&existingItem, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusNotFound, "Item not found", err)
	}

	var updatedItem models.FirstStockItems
	if err := c.BodyParser(&updatedItem); err != nil {
		return helpers.JSONResponse(c, http.StatusBadRequest, "Invalid input", err)
	}

	// Rollback stok lama
	if err := services.ReduceProductStock(db, existingItem.ProductId, existingItem.Qty); err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to rollback product stock", err)
	}

	// Tambah stok baru
	if err := services.AddProductStock(db, updatedItem.ProductId, updatedItem.Qty); err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to add product stock", err)
	}

	// Update item
	existingItem.ProductId = updatedItem.ProductId
	existingItem.Qty = updatedItem.Qty
	existingItem.Price = updatedItem.Price
	existingItem.SubTotal = updatedItem.Price * updatedItem.Qty

	if err := db.Save(&existingItem).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to update FirstStock item", err)
	}

	// Update harga produk jika harga item lebih tinggi
	if err := services.UpdateProductPriceIfHigher(db, updatedItem.ProductId, updatedItem.Price); err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to update product price", err)
	}

	go func() {
		if err := RecalculateTotalFirstStock(db, existingItem.FirstStockId); err != nil {
			fmt.Printf("Failed to recalculate total FirstStock asynchronously: %v\n", err)
		}
	}()

	return helpers.JSONResponse(c, http.StatusOK, "Item updated successfully", existingItem)
}

// Delete FirstStockItem
func DeleteFirstStockItem(c *fiber.Ctx) error {
	db := configs.DB
	id := c.Params("id")

	var item models.FirstStockItems
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusNotFound, "Item not found", err)
	}

	// Subtract stok
	if err := services.ReduceProductStock(db, item.ProductId, item.Qty); err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to rollback product stock", err)
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

	return helpers.JSONResponse(c, http.StatusOK, "Item deleted successfully", item)
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
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid month format. Use YYYY-MM", nil)
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
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get FirstStocks failed", err.Error())
	}

	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "FirstStocks retrieved successfully", search, int(total), page, int(totalPages), int(limit), FirstStocks)
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
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to get items", err)
	}

	return helpers.JSONResponse(c, http.StatusOK, "Items retrieved successfully", FirstStockItems)
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
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to get first_stock", err)
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
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to get FirstStock items", err)
	}

	// Format tanggal pembelian ke dd-mm-yyyy
	formattedDate := first_stock.FirstStockDate.Format("02-01-2006")

	return JSONFirstStockWithItemsResponse(c, http.StatusOK, "FirstStock retrieved successfully", first_stockID, first_stock.Description, formattedDate, first_stock.TotalFirstStock, formattedDate, items)
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
		return helpers.JSONResponse(c, http.StatusBadRequest, "Invalid request body", err)
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
		return helpers.JSONResponse(c, http.StatusBadRequest, "Validation failed for first stock header input", err)
	}
	for _, item := range req.FirstStockItems {
		if err = helpers.ValidateStruct(item); err != nil {
			return helpers.JSONResponse(c, http.StatusBadRequest, "Validation failed for one or more first stock items", err)
		}
		// Validasi manual tambahan karena tag required di-relax
		if item.ProductId == "" || item.UnitId == "" {
			return helpers.JSONResponse(c, http.StatusBadRequest, "product_id and unit_id are required for all items", nil)
		}
		if item.Qty <= 0 {
			return helpers.JSONResponse(c, http.StatusBadRequest, "qty must be greater than 0", nil)
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
			return helpers.JSONResponse(c, http.StatusBadRequest, "Invalid first_stock_date format. Please use `YYYY-MM-DD`.", err.Error())
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
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to begin database transaction", tx.Error)
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
			tx.Rollback()
			return helpers.JSONResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid expired_date format for product %s. Please use `YYYY-MM-DD`.", reqItem.ProductId), err)
		}

		var product models.Product
		err = tx.Where("id = ? AND branch_id = ?", reqItem.ProductId, firstStockHeader.BranchID).First(&product).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return helpers.JSONResponse(c, http.StatusNotFound, fmt.Sprintf("Product with ID %s not found in branch %s", reqItem.ProductId, firstStockHeader.BranchID), err)
			}
			return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to retrieve product details", err)
		}

		// Mendapatkan detail unit (sesuai unit_id yang diinput)
		var unit models.Unit
		err = tx.Where("id = ?", reqItem.UnitId).First(&unit).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return helpers.JSONResponse(c, http.StatusNotFound, fmt.Sprintf("Unit with ID %s not found", reqItem.UnitId), err)
			}
			return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to retrieve unit details", err)
		}

		// --- Logika Konversi Satuan ---
		var conversionValue int = 1
		if reqItem.UnitId != product.UnitId { // Jika unit input berbeda dengan unit dasar produk
			var unitConversion models.UnitConversion
			err = tx.Where("product_id = ? AND init_id = ? AND final_id = ? AND branch_id = ?",
				reqItem.ProductId,
				reqItem.UnitId, // Unit yang diinput
				product.UnitId, // Unit dasar produk
				firstStockHeader.BranchID,
			).First(&unitConversion).Error

			if err != nil {
				if err == gorm.ErrRecordNotFound {
					// Jika tidak ada konversi yang didefinisikan, asumsikan 1:1 atau unit dasar.
					// Anda bisa menambahkan error di sini jika konversi mutlak diperlukan.
					// Saat ini dibiarkan conversionValue = 1
				} else {
					tx.Rollback()
					return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to retrieve unit conversion details", err)
				}
			} else {
				conversionValue = unitConversion.ValueConv
			}
		}
		actualQtyToAdd := reqItem.Qty * conversionValue // Kuantitas aktual dalam satuan dasar
		// --- Akhir Logika Konversi Satuan ---

		// Harga untuk First Stock Items diambil dari PurchasePrice produk
		// Ini merepresentasikan "nilai" dari stok yang masuk, bukan biaya.
		itemPrice := product.PurchasePrice         // Harga beli per unit dasar produk
		itemSubTotal := itemPrice * actualQtyToAdd // SubTotal berdasarkan harga beli dan kuantitas aktual

		firstStockItemDB := models.FirstStockItems{
			ID:           helpers.GenerateID("FSI"), // ID untuk First Stock Item
			FirstStockId: firstStockHeader.ID,
			ProductId:    reqItem.ProductId,
			Price:        itemPrice,    // Price dari PurchasePrice produk
			Qty:          reqItem.Qty,  // Qty yang diinput (dalam unit yang diinput)
			SubTotal:     itemSubTotal, // SubTotal yang dihitung
			ExpiredDate:  parsedExpiredDate,
		}
		firstStockItemsToCreate = append(firstStockItemsToCreate, firstStockItemDB)

		// --- Siapkan data untuk respons ---
		firstStockItemResp := models.FirstStockItemResponse{
			ID:          firstStockItemDB.ID,
			ProductID:   firstStockItemDB.ProductId,
			ProductName: product.Name,
			UnitID:      reqItem.UnitId, // Unit yang diinput
			UnitName:    unit.Name,
			Price:       firstStockItemDB.Price,
			Qty:         firstStockItemDB.Qty,
			SubTotal:    firstStockItemDB.SubTotal,
			ExpiredDate: parsedExpiredDate.Format("02 January 2006"), // Format tanggal
		}
		firstStockItemsForResponse = append(firstStockItemsForResponse, firstStockItemResp)
		// --- Akhir persiapan data respons ---

		// --- Tambah stok dan cek/update expired_date ---
		updates := map[string]interface{}{
			"stock": product.Stock + actualQtyToAdd, // Tambahkan stok aktual (dalam satuan dasar)
		}

		// Jika ExpiredDate stok baru lebih awal dari yang sudah ada di master produk, update.
		if parsedExpiredDate.Before(product.ExpiredDate) {
			updates["expired_date"] = parsedExpiredDate
		}

		err = tx.Model(&models.Product{}).Where("id = ?", product.ID).Updates(updates).Error
		if err != nil {
			tx.Rollback()
			return helpers.JSONResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to update product details (stock/expired_date) for product %s", product.Name), err)
		}

		calculatedTotalFirstStock += itemSubTotal // Ini adalah nilai total stok yang dimasukkan
	}

	firstStockHeader.TotalFirstStock = calculatedTotalFirstStock

	// Simpan data FirstStocks
	err = tx.Create(&firstStockHeader).Error
	if err != nil {
		tx.Rollback()
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to create first stock entry", err)
	}

	// Simpan FirstStockItems dalam batch
	err = tx.CreateInBatches(&firstStockItemsToCreate, len(firstStockItemsToCreate)).Error
	if err != nil {
		tx.Rollback()
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to create first stock items", err)
	}

	// PENTING: TransactionReports dan DailyProfitReport TIDAK relevan untuk First Stock
	// Karena ini bukan transaksi finansial atau penjualan/pembelian berbiaya,
	// bagian untuk membuat TransactionReports atau mengupdate DailyProfitReport dihapus.

	// Cek `subscription_type` jika type nya adalah `quota`
	// Asumsi: First Stock TIDAK mengurangi kuota transaksi.
	// Jika first stock harus mengurangi kuota (misal, setiap entri dianggap transaksi),
	// Anda bisa menambahkan logika pengurangan kuota di sini.
	if subscriptionType == "quota" {
		var branch models.Branch
		err = tx.Where("id = ?", firstStockHeader.BranchID).First(&branch).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return helpers.JSONResponse(c, http.StatusNotFound, fmt.Sprintf("Branch with ID %s not found", firstStockHeader.BranchID), err)
			}
			return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to retrieve branch details for quota check", err)
		}
		// Logika pengurangan kuota dibiarkan kosong di sini, karena first stock tidak mengurangi kuota.
	}

	err = tx.Commit().Error
	if err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to commit database transaction for first stock", err)
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

	return helpers.JSONResponse(c, http.StatusCreated, "First stock transaction created successfully", response)
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
