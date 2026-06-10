package handlers

import (
	errors "errors"
	fmt "fmt"
	math "math"
	strconv "strconv"
	strings "strings"
	time "time"

	configs "apotek-clean/configs"
	helpers "apotek-clean/helpers"
	models "apotek-clean/models"
	services "apotek-clean/services"
	reports "apotek-clean/services/reports"
	fiber "github.com/gofiber/fiber/v2"
	gorm "gorm.io/gorm"
)

func syncPurchaseItemPostSave(db *gorm.DB, purchaseID, productID string, qty, price int) error {
	if err := services.AddProductStock(db, productID, qty); err != nil {
		return err
	}

	if err := services.UpdateProductPriceIfHigher(db, productID, price); err != nil {
		return err
	}

	return reports.RecalculateTotalPurchase(db, purchaseID)
}

func recalculatePurchaseAfterItemChange(db *gorm.DB, purchaseID string) error {
	return reports.RecalculateTotalPurchase(db, purchaseID)
}

type preparedPurchaseTransactionItem struct {
	purchaseItem  models.PurchaseItems
	responseItem  models.PurchaseItemResponse
	productID     string
	productName   string
	productUpdate map[string]interface{}
	subTotal      int
}

func createPurchaseTransactionReport(tx *gorm.DB, purchase models.Purchases, nowWIB time.Time) error {
	transactionReportID := helpers.GenerateID("TRX")
	transactionReport := models.TransactionReports{
		ID:              transactionReportID,
		TransactionType: models.Purchase,
		UserID:          purchase.UserID,
		BranchID:        purchase.BranchID,
		Total:           purchase.TotalPurchase,
		Payment:         purchase.Payment,
		CreatedAt:       nowWIB,
		UpdatedAt:       nowWIB,
	}
	return tx.Create(&transactionReport).Error
}

func rollbackWithJSON(c *fiber.Ctx, tx *gorm.DB, status int, message string, err error) error {
	tx.Rollback()
	return helpers.JSONResponse(c, status, message, err)
}

func applyPurchaseQuotaIfNeeded(tx *gorm.DB, subscriptionType string, branchID string) error {
	if subscriptionType != "quota" {
		return nil
	}

	var branch models.Branch
	err := tx.Where("id = ?", branchID).First(&branch).Error
	if err != nil {
		return err
	}

	if branch.Quota <= 0 {
		return errors.New("quota exceeded")
	}

	branch.Quota -= 1
	return tx.Save(&branch).Error
}

func preparePurchaseTransactionItem(purchaseID string, itemInput models.PurchaseItemInput, lookup services.PurchaseItemLookupResult, parsedExpiredDate time.Time) preparedPurchaseTransactionItem {
	preparedValues := services.PreparePurchaseItemValues(itemInput.Qty, itemInput.Price, lookup.ConversionValue)
	purchaseItem := services.BuildPurchaseItemModel(helpers.GenerateID("PIT"), services.PurchaseItemModelParams{
		PurchaseID:  purchaseID,
		ProductID:   itemInput.ProductId,
		UnitID:      itemInput.UnitId,
		Price:       preparedValues.ItemPrice,
		Qty:         itemInput.Qty,
		SubTotal:    preparedValues.ItemSubTotal,
		ExpiredDate: parsedExpiredDate,
	})
	responseItem := services.BuildPurchaseItemResponse(services.PurchaseItemResponseParams{
		Item:        purchaseItem,
		ProductName: lookup.Product.Name,
		UnitName:    lookup.Unit.Name,
		ExpiredDate: parsedExpiredDate,
	})
	return preparedPurchaseTransactionItem{
		purchaseItem:  purchaseItem,
		responseItem:  responseItem,
		productID:     lookup.Product.ID,
		productName:   lookup.Product.Name,
		productUpdate: services.BuildPurchasedProductUpdates(lookup.Product, preparedValues.ActualQtyToAdd, parsedExpiredDate),
		subTotal:      preparedValues.ItemSubTotal,
	}
}

// CreatePurchase Function is using to create new purchase
func CreatePurchase(c *fiber.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	db := configs.DB

	// Ambil informasi dari token
	branchID, _ := services.GetBranchID(c)
	userID, _ := services.GetUserID(c)
	generatedID := helpers.GenerateID("PUR")

	// Ambil input dari body
	var input models.PurchaseInput
	if err := c.BodyParser(&input); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Input pembelian tidak valid", nil)
	}

	// Parse tanggal
	layout := "2006-01-02" // format harus YYYY-MM-DD
	parsedDate, err := time.Parse(layout, input.PurchaseDate)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Format tanggal tidak valid. Gunakan YYYY-MM-DD", nil)
	}

	// Map ke struct model
	purchase := models.Purchases{
		ID:            generatedID,
		SupplierId:    input.SupplierId,
		BranchID:      branchID,
		UserID:        userID,
		PurchaseDate:  parsedDate,
		TotalPurchase: 0,
		CreatedAt:     nowWIB,
		UpdatedAt:     nowWIB,
	}

	// Simpan purchase
	if err := db.Create(&purchase).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal membuat pembelian", err)
	}

	// Buat laporan
	if err := reports.SyncPurchaseReport(db, purchase); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal menyinkronkan laporan pembelian", err)
	}

	_ = reports.AutoCleanupPurchases(db)

	return helpers.JSONResponse(c, fiber.StatusOK, "Pembelian berhasil dibuat", purchase)
}

// UpdatePurchase Function is using to update purchase
func UpdatePurchase(c *fiber.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	db := configs.DB
	id := c.Params("id")

	// Cari data purchase lama
	var purchase models.Purchases
	if err := db.First(&purchase, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Pembelian tidak ditemukan", nil)
	}

	// 🔁 Panggil reusable function untuk validasi 1 jam
	if err := services.EnsurePurchaseEditable(db, purchase.ID); err != nil {
		if errors.Is(err, services.ErrDataExpiredToEdit) {
			return helpers.JSONResponse(c, fiber.StatusForbidden, err.Error(), nil)
		}
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil timestamp pembelian", err)
	}

	// Gunakan struct input
	var input models.PurchaseInput
	if err := c.BodyParser(&input); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Input pembelian tidak valid", err)
	}

	// Cek dan update SupplierID
	if input.SupplierId != "" {
		purchase.SupplierId = input.SupplierId
	}

	// Cek dan update PurchaseDate
	if input.PurchaseDate != "" {
		layout := "2006-01-02"
		parsedDate, err := time.Parse(layout, input.PurchaseDate)
		if err != nil {
			return helpers.JSONResponse(c, fiber.StatusBadRequest, "Format tanggal tidak valid. Gunakan YYYY-MM-DD", err)
		}
		purchase.PurchaseDate = parsedDate
	}

	// Cek dan update Payment
	if input.Payment != "" {
		purchase.Payment = models.PaymentStatus(input.Payment)
	}

	purchase.UpdatedAt = nowWIB

	// Hitung ulang total dari purchase items
	var items []models.PurchaseItems
	if err := db.Where("purchase_id = ?", id).Find(&items).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil item pembelian", err)
	}

	purchase.TotalPurchase = services.SumPurchaseItems(items)

	// Simpan perubahan
	if err := db.Save(&purchase).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui pembelian", err)
	}

	// Sync report
	if err := reports.SyncPurchaseReport(db, purchase); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal menyinkronkan laporan pembelian", err)
	}

	_ = reports.AutoCleanupPurchases(db)

	return helpers.JSONResponse(c, fiber.StatusOK, "Pembelian berhasil diperbarui", purchase)
}

// DeletePurchase Function
func DeletePurchase(c *fiber.Ctx) error {
	db := configs.DB
	id := c.Params("id")

	// Ambil purchase
	var purchase models.Purchases
	if err := db.First(&purchase, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Pembelian tidak ditemukan", nil)
	}

	// 🔁 Panggil reusable function untuk validasi 1 jam
	if err := services.EnsurePurchaseEditable(db, purchase.ID); err != nil {
		if errors.Is(err, services.ErrDataExpiredToEdit) {
			return helpers.JSONResponse(c, fiber.StatusForbidden, err.Error(), nil)
		}
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil timestamp pembelian", err)
	}

	// Ambil item-item dan rollback stok
	var items []models.PurchaseItems
	if err := db.Where("purchase_id = ?", id).Find(&items).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil item pembelian", err)
	}

	for _, item := range items {
		// Rollback stok ke produk
		if err := services.ReduceProductStock(db, item.ProductId, item.Qty); err != nil {
			return helpers.JSONResponse(c, fiber.StatusInternalServerError, fmt.Sprintf("Gagal rollback stok untuk product ID %s", item.ProductId), err)
		}
	}

	// Hapus semua item dari pembelian
	if err := db.Where("purchase_id = ?", id).Delete(&models.PurchaseItems{}).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal menghapus item pembelian", err)
	}

	// Hapus laporan transaksi terkait
	if err := db.Where("id = ? AND transaction_type = ?", purchase.ID, models.Purchase).Delete(&models.TransactionReports{}).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal menghapus laporan transaksi", err)
	}

	// Hapus purchase
	if err := db.Delete(&purchase).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal menghapus pembelian", err)
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Pembelian berhasil dihapus", purchase)
}

// CreatePurchaseItem Function is using to create new purchase item
func CreatePurchaseItem(c *fiber.Ctx) error {
	db := configs.DB
	var item models.PurchaseItems

	if err := c.BodyParser(&item); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Input item pembelian tidak valid", err)
	}

	// 🔁 Panggil reusable function untuk validasi 1 jam
	if err := services.EnsurePurchaseEditable(db, item.PurchaseId); err != nil {
		if errors.Is(err, services.ErrDataExpiredToEdit) {
			return helpers.JSONResponse(c, fiber.StatusForbidden, err.Error(), nil)
		}
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil timestamp pembelian", err)
	}

	// Cek apakah item dengan purchase_id dan product_id sudah ada
	var existing models.PurchaseItems
	err := db.Where("purchase_id = ? AND product_id = ?", item.PurchaseId, item.ProductId).First(&existing).Error
	if err == nil {
		// Sudah ada: update qty dan sub_total
		existing.Qty += item.Qty
		existing.SubTotal = existing.Qty * existing.Price // asumsi pakai harga awal

		if err := db.Save(&existing).Error; err != nil {
			return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui item pembelian yang sudah ada", err)
		}

		if err := syncPurchaseItemPostSave(db, item.PurchaseId, item.ProductId, item.Qty, item.Price); err != nil {
			return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal menyinkronkan efek item pembelian", err)
		}

		return helpers.JSONResponse(c, fiber.StatusOK, "Item pembelian berhasil diperbarui", existing)

	} else if err != gorm.ErrRecordNotFound {
		// Error selain record not found
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal memeriksa item pembelian yang sudah ada", err)
	}

	// Data belum ada, buat item baru
	if item.ID == "" {
		item.ID = helpers.GenerateID("PIT")
	}
	item.SubTotal = item.Qty * item.Price

	if err := db.Create(&item).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal membuat item pembelian", err)
	}

	if err := syncPurchaseItemPostSave(db, item.PurchaseId, item.ProductId, item.Qty, item.Price); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal menyinkronkan efek item pembelian", err)
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Item pembelian berhasil ditambahkan", item)
}

// Update PurchaseItem is using to update purchase
func UpdatePurchaseItem(c *fiber.Ctx) error {
	db := configs.DB
	id := c.Params("id")

	var existingItem models.PurchaseItems
	if err := db.First(&existingItem, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Item tidak ditemukan", nil)
	}

	// 🔁 Panggil reusable function untuk validasi 1 jam
	if err := services.EnsurePurchaseEditable(db, existingItem.PurchaseId); err != nil {
		if errors.Is(err, services.ErrDataExpiredToEdit) {
			return helpers.JSONResponse(c, fiber.StatusForbidden, err.Error(), nil)
		}
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil timestamp pembelian", err)
	}

	var updatedItem models.PurchaseItems
	if err := c.BodyParser(&updatedItem); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Input item pembelian tidak valid", err)
	}

	// Rollback stok lama
	if err := services.ReduceProductStock(db, existingItem.ProductId, existingItem.Qty); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal rollback stok lama", err)
	}

	// Tambah stok baru
	if err := services.AddProductStock(db, updatedItem.ProductId, updatedItem.Qty); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal menambahkan stok baru", err)
	}

	// Update item
	existingItem.ProductId = updatedItem.ProductId
	existingItem.Qty = updatedItem.Qty
	existingItem.Price = updatedItem.Price
	existingItem.SubTotal = updatedItem.Price * updatedItem.Qty

	if err := db.Save(&existingItem).Error; err != nil {
		// return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui item pembelian", err)
	}

	if err := services.UpdateProductPriceIfHigher(db, updatedItem.ProductId, updatedItem.Price); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui harga produk", err)
	}

	if err := recalculatePurchaseAfterItemChange(db, existingItem.PurchaseId); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal menghitung ulang total pembelian", err)
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Item pembelian berhasil diperbarui", existingItem)
}

// Delete PurchaseItem is using to delete purchase
func DeletePurchaseItem(c *fiber.Ctx) error {
	db := configs.DB
	id := c.Params("id")

	var item models.PurchaseItems
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Item tidak ditemukan", nil)
	}

	// 🔁 Panggil reusable function untuk validasi 1 jam
	if err := services.EnsurePurchaseEditable(db, item.PurchaseId); err != nil {
		if errors.Is(err, services.ErrDataExpiredToEdit) {
			return helpers.JSONResponse(c, fiber.StatusForbidden, err.Error(), nil)
		}
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil timestamp pembelian", err)
	}

	// Subtract stok
	if err := services.ReduceProductStock(db, item.ProductId, item.Qty); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengurangi stok produk", err)
	}

	// Hapus item
	if err := db.Delete(&item).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal menghapus item pembelian", err)
	}

	if err := recalculatePurchaseAfterItemChange(db, item.PurchaseId); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal menghitung ulang total pembelian", err)
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Item pembelian berhasil dihapus", item)
}

// Get All Purchases tampilkan semua purchase
func GetAllPurchases(c *fiber.Ctx) error {
	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	branchID, _ := services.GetBranchID(c)

	// Ambil parameter page dan search dari query URL
	pageParam := c.Query("page")
	search := strings.TrimSpace(c.Query("search"))

	// Konversi page ke int, default ke 1 jika tidak valid
	page := 1
	if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
		page = p
	}

	limit := 10                  // Tetapkan limit ke 10 data per halaman
	offset := (page - 1) * limit // Hitung offset berdasarkan halaman dan limit

	month := strings.TrimSpace(c.Query("month"))

	// Jika month kosong, isi dengan bulan ini (format YYYY-MM)
	if month == "" {
		month = nowWIB.Format("2006-01")
	}

	var purchases []models.AllPurchases
	var total int64

	// Mulai bangun query
	query := configs.DB.Table("purchases pur").
		Select("pur.id, pur.supplier_id, sup.name AS supplier_name, pur.purchase_date, pur.total_purchase, pur.payment").
		Joins("LEFT JOIN suppliers sup ON sup.id = pur.supplier_id").
		Where("pur.branch_id = ? AND pur.total_purchase > 0", branchID)

	// Filter bulan
	startDate, err := time.Parse("2006-01", month)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid month format", err)
	}
	endDate := startDate.AddDate(0, 1, 0)
	query = query.Where("pur.purchase_date >= ? AND pur.purchase_date < ?", startDate, endDate)

	// Filter search jika ada
	if search != "" {
		search = strings.ToLower(search)
		query = query.Where("LOWER(sup.name) LIKE ?", "%"+search+"%")
	}

	// Hitung total
	if err := query.Count(&total).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get purchase failed", err)
	}

	// Ambil data paginasi
	if err := query.Order("pur.created_at DESC").Offset(offset).Limit(limit).Scan(&purchases).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get purchases failed", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// Buat slice baru untuk menampung data yang sudah diformat
	var formattedPurchasesData []models.PurchaseDetailResponse
	for _, purchase := range purchases {
		formattedPurchasesData = append(formattedPurchasesData, models.PurchaseDetailResponse{
			ID:            purchase.ID,
			SupplierId:    purchase.SupplierId,
			SupplierName:  purchase.SupplierName,
			PurchaseDate:  helpers.FormatIndonesianDate(purchase.PurchaseDate), // Format tanggal di sini
			TotalPurchase: purchase.TotalPurchase,
			Payment:       string(purchase.Payment),
		})
	}

	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Data pembelian berhasil diambil", search, int(total), page, totalPages, limit, formattedPurchasesData)
}

// GetAllPurchaseItems tampilkan semua item berdasarkan purchase_id tanpa pagination
func GetAllPurchaseItems(c *fiber.Ctx) error {
	// Get purchase id dari param
	purchaseID := c.Params("id")

	// Parsing body JSON ke struct
	var PurchaseItems []models.AllPurchaseItems

	// Query dasar
	query := configs.DB.Table("purchase_items pit").
		Select("pit.id, pit.purchase_id, pit.product_id, pro.name AS product_name, pit.price, pit.qty, pro.unit_id, un.name AS unit_name, pit.sub_total, pit.expired_date").
		Joins("LEFT JOIN products pro ON pro.id = pit.product_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("pit.purchase_id = ?", purchaseID).
		Order("pro.name ASC")

	// Eksekusi query
	if err := query.Scan(&PurchaseItems).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil item pembelian", err)
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Data item pembelian berhasil diambil", PurchaseItems)
}

// GetPurchaseWithItems menampilkan satu purchase beserta semua item-nya
func GetPurchaseWithItems(c *fiber.Ctx) error {
	db := configs.DB

	// Ambil ID pembelian dari parameter URL
	purchaseID := c.Params("id")

	// Struct untuk data utama purchase
	var purchase models.AllPurchases

	// Ambil data purchase dengan LEFT JOIN ke suppliers
	err := db.Table("purchases pur").
		Select("pur.id, pur.supplier_id, sup.name AS supplier_name, pur.purchase_date, pur.total_purchase, pur.payment").
		Joins("LEFT JOIN suppliers sup ON sup.id = pur.supplier_id").
		Where("pur.id = ?", purchaseID).
		Scan(&purchase).Error

	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to get purchase", err)
	}

	// Ambil item pembelian terkait
	var items []models.AllPurchaseItems
	err = db.Table("purchase_items pit").
		Select("pit.id, pit.purchase_id, pit.product_id, pro.name AS product_name, pit.unit_id AS unit_id, un.name AS unit_name, pit.price, pit.qty, pit.sub_total, pit.expired_date").
		Joins("LEFT JOIN products pro ON pro.id = pit.product_id").
		Joins("LEFT JOIN units un ON un.id = pit.unit_id").
		Where("pit.purchase_id = ?", purchaseID).
		Order("pro.name ASC").
		Scan(&items).Error

	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil item pembelian", err)
	}

	// Buat slice baru untuk menampung data yang sudah diformat
	var formatedPurchaseItems []models.FormatedPurchaseItems
	for _, purItems := range items {
		formatedPurchaseItems = append(formatedPurchaseItems, models.FormatedPurchaseItems{
			ID:          purItems.ID,
			ProductId:   purItems.ProductId,
			ProductName: purItems.ProductName,
			UnitId:      purItems.UnitId,
			UnitName:    purItems.UnitName,
			Price:       purItems.Price,
			Qty:         purItems.Qty,
			SubTotal:    purItems.SubTotal,
			ExpiredDate: helpers.FormatIndonesianDate(purItems.ExpiredDate), // Format tanggal di sini
		})
	}

	// Format tanggal secara manual untuk respons ini
	// Menggunakan helper FormatIndonesianDate yang sudah kita buat
	formattedPurchaseDate := helpers.FormatIndonesianDate(purchase.PurchaseDate)

	// Buat objek respons menggunakan struct PurchaseItemResponse yang baru
	// dan isi field-fieldnya
	responseDetail := models.PurchaseDetailWithItemsResponse{
		ID:            purchase.ID,
		SupplierId:    purchase.SupplierId,
		SupplierName:  purchase.SupplierName,
		PurchaseDate:  formattedPurchaseDate, // Gunakan tanggal yang sudah diformat
		TotalPurchase: purchase.TotalPurchase,
		Payment:       string(purchase.Payment),
		Items:         formatedPurchaseItems,
	}

	// Panggil JSONResponse yang sudah ada, meneruskan PurchaseItemResponse sebagai 'data'
	return helpers.JSONResponse(c, fiber.StatusOK, "Data pembelian berhasil diambil", responseDetail)
}

// CreatePurchaseTransaction controller
func CreatePurchaseTransaction(c *fiber.Ctx) error {
	nowWIB := time.Now().In(configs.Location)

	subscriptionType, _ := services.GetClaimsToken(c, "subscription_type")
	branchID, _ := services.GetBranchID(c)
	userID, _ := services.GetUserID(c)

	db := configs.DB
	var req models.PurchaseTransactionRequest
	err := c.BodyParser(&req)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Body permintaan tidak valid", err)
	}

	if req.Purchase.Payment == "" {
		req.Purchase.Payment = "paid_by_cash"
	}

	req.Purchase.UserID = userID
	req.Purchase.BranchID = branchID

	err = helpers.ValidateStruct(req.Purchase)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Validasi input pembelian gagal", err)
	}

	for _, item := range req.PurchaseItems {
		err = helpers.ValidateStruct(item)
		if err != nil {
			return helpers.JSONResponse(c, fiber.StatusBadRequest, "Validasi satu atau lebih item pembelian gagal", err)
		}
	}

	purchaseDate, err := services.ParsePurchaseDate(req.Purchase.PurchaseDate, nowWIB)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid purchase_date format. Please use `YYYY-MM-DD`.", err)
	}

	purchase := models.Purchases{
		SupplierId:   req.Purchase.SupplierId,
		PurchaseDate: purchaseDate,
		BranchID:     req.Purchase.BranchID,
		Payment:      req.Purchase.Payment,
		UserID:       req.Purchase.UserID,
	}

	// --- Proses Penyimpanan Data ---
	tx := db.Begin()
	if tx.Error != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal memulai transaksi database", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	purchaseID := helpers.GenerateID("PUR")
	purchase.ID = purchaseID
	purchase.CreatedAt = nowWIB
	purchase.UpdatedAt = nowWIB

	var calculatedTotalPurchase int
	var purchaseItemsToCreate []models.PurchaseItems
	var purchaseItemsForResponse []models.PurchaseItemResponse // <--- Slice baru untuk data respons

	// Mendapatkan nama supplier (di luar loop item untuk efisiensi)
	supplier, err := services.LookupPurchaseSupplier(tx, req.Purchase.SupplierId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return rollbackWithJSON(c, tx, fiber.StatusNotFound, fmt.Sprintf("Supplier with ID %s not found", req.Purchase.SupplierId), err)
		}
		return rollbackWithJSON(c, tx, fiber.StatusInternalServerError, "Gagal mengambil detail supplier", err)
	}

	// var stockTracksToCreate []models.StockTracks

	for i := range req.PurchaseItems {
		parsedExpiredDate, err := time.Parse("2006-01-02", req.PurchaseItems[i].ExpiredDate)
		if err != nil {
			tx.Rollback()
			return helpers.JSONResponse(c, fiber.StatusBadRequest, fmt.Sprintf("Invalid expired_date format for product %s. Please use `YYYY-MM-DD`.", req.PurchaseItems[i].ProductId), err)
		}

		lookup, err := services.LookupPurchaseItemDependencies(tx, purchase.BranchID, req.PurchaseItems[i].ProductId, req.PurchaseItems[i].UnitId)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				if tx.Where("id = ?", req.PurchaseItems[i].ProductId).First(&models.Product{}).Error == nil {
					return rollbackWithJSON(c, tx, fiber.StatusNotFound, fmt.Sprintf("Unit with ID %s not found", req.PurchaseItems[i].UnitId), err)
				}
				return rollbackWithJSON(c, tx, fiber.StatusNotFound, fmt.Sprintf("Product with ID %s not found", req.PurchaseItems[i].ProductId), err)
			}
			return rollbackWithJSON(c, tx, fiber.StatusInternalServerError, "Gagal mengambil dependensi item pembelian", err)
		}

		preparedTransactionItem := preparePurchaseTransactionItem(purchaseID, req.PurchaseItems[i], lookup, parsedExpiredDate)
		purchaseItemsToCreate = append(purchaseItemsToCreate, preparedTransactionItem.purchaseItem)
		purchaseItemsForResponse = append(purchaseItemsForResponse, preparedTransactionItem.responseItem)

		err = tx.Model(&models.Product{}).Where("id = ?", preparedTransactionItem.productID).Updates(preparedTransactionItem.productUpdate).Error
		if err != nil {
			return rollbackWithJSON(c, tx, fiber.StatusInternalServerError, fmt.Sprintf("Gagal memperbarui detail produk (stok/expired_date) untuk produk %s", preparedTransactionItem.productName), err)
		}
		calculatedTotalPurchase += preparedTransactionItem.subTotal
	}

	purchase.TotalPurchase = calculatedTotalPurchase

	err = tx.Create(&purchase).Error
	if err != nil {
		return rollbackWithJSON(c, tx, fiber.StatusInternalServerError, "Gagal membuat pembelian", err)
	}

	err = tx.CreateInBatches(&purchaseItemsToCreate, len(purchaseItemsToCreate)).Error
	if err != nil {
		return rollbackWithJSON(c, tx, fiber.StatusInternalServerError, "Gagal membuat item pembelian", err)
	}

	err = createPurchaseTransactionReport(tx, purchase, nowWIB)
	if err != nil {
		return rollbackWithJSON(c, tx, fiber.StatusInternalServerError, "Gagal membuat laporan transaksi pembelian", err)
	}

	err = applyPurchaseQuotaIfNeeded(tx, subscriptionType, req.Purchase.BranchID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return rollbackWithJSON(c, tx, fiber.StatusNotFound, fmt.Sprintf("Branch with ID %s not found", req.Purchase.BranchID), err)
		}
		if err.Error() == "quota exceeded" {
			return rollbackWithJSON(c, tx, fiber.StatusBadRequest, "Tidak ada kuota tersedia untuk cabang", err)
		}
		return rollbackWithJSON(c, tx, fiber.StatusInternalServerError, "Gagal menerapkan kuota untuk pembelian", err)
	}

	err = tx.Commit().Error
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal melakukan commit transaksi database", err)
	}

	// --- Akhir: Mengkonstruksi Objek Respon ---
	response := services.BuildPurchaseTransactionResponse(purchase, supplier.Name, purchaseDate, purchaseItemsForResponse)
	// --- Akhir Mengkonstruksi Objek Respon ---

	return helpers.JSONResponse(c, fiber.StatusOK, "Transaksi pembelian berhasil dibuat", response)
}

// GetFixedPrice menghitung harga produk setelah konversi satuan
// Endpoint ini akan dipanggil dengan query parameters: product_id, init_id, final_id
// Contoh: GET /api/fixed-price?product_id=PRD123&init_id=UNT_BOX&final_id=UNT_PCS
func GetFixedPrice(c *fiber.Ctx) error {
	db := configs.DB
	var req models.GetFixedPriceRequest

	// Parsing query parameters
	if err := c.QueryParser(&req); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Query parameter tidak valid", err)
	}

	// Get BranchID from token (jika harga dan konversi spesifik per cabang)
	// Jika konversi satuan dan harga tidak spesifik cabang, Anda bisa hapus baris ini dan filter branch_id di query
	branchID, _ := services.GetBranchID(c)
	if branchID == "" {
		return helpers.JSONResponse(c, fiber.StatusUnauthorized, "Branch ID tidak ditemukan di token. Unauthorized.", nil)
	}

	// antangin sachet 12
	// tolak angin 12
	// amlodipin strip 10 box
	// grantusif 10 box
	// asamefenamat 10 box

	// --- 1. Dapatkan Product dan PurchasePrice-nya ---
	var product models.Product
	err := db.Where("id = ? AND branch_id = ?", req.ProductID, branchID).First(&product).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return helpers.JSONResponse(c, fiber.StatusNotFound, fmt.Sprintf("Product with ID %s not found in branch %s", req.ProductID, branchID), err)
		}
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil detail produk", err)
	}

	// --- 2. Dapatkan UnitConversion Value ---
	var conversionValue int = 1 // Default to 1 if no conversion or no explicit FinalID
	var unitConversion models.UnitConversion

	// Asumsi: jika final_id dari request adalah sama dengan init_id, atau
	// jika init_id sudah merupakan unit dasar, maka value_conv adalah 1.
	// Jika req.InitID sama dengan req.FinalID, maka konversinya 1.
	// Jika UnitId di Product adalah satuan dasar, maka FinalId di UnitConversion harus merujuk ke UnitId di Product.
	// Kita akan mencari konversi dari init_id ke final_id yang diberikan di parameter.

	// Hanya cari konversi jika init_id tidak sama dengan final_id,
	// dan juga init_id tidak sama dengan unit default produk (jika diasumsikan sebagai satuan dasar).
	if req.InitID != req.FinalID { // Konversi hanya diperlukan jika satuan awal dan akhir berbeda
		err = db.Where("product_id = ? AND init_id = ? AND final_id = ? AND branch_id = ?",
			req.ProductID,
			req.InitID,
			req.FinalID,
			branchID,
		).First(&unitConversion).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// Jika konversi spesifik tidak ditemukan, kita bisa memilih untuk:
				// a) Mengembalikan error (lebih ketat)
				// b) Mengasumsikan value_conv adalah 1 (lebih longgar, mungkin berarti unit_id == final_id secara implisit)
				// Untuk endpoint ini, saya akan mengembalikan error karena permintaan Anda eksplisit.
				return helpers.JSONResponse(c, fiber.StatusNotFound, fmt.Sprintf("Unit conversion from %s to %s for product %s not found in branch %s", req.InitID, req.FinalID, req.ProductID, branchID), err)
			}
			return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil detail konversi satuan", err)
		}
		conversionValue = unitConversion.ValueConv
	}

	// --- 3. Hitung FixPrice ---
	fixPrice := product.PurchasePrice * conversionValue

	// --- 4. Buat Respon ---
	response := models.FixedPriceResponse{
		FixPrice: fixPrice,
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Harga tetap berhasil dihitung", response)
}

// GetUnitsByProductIdRequest merepresentasikan body request untuk endpoint GET ini
type GetUnitsByProductIdRequest struct {
	ProductID string `json:"product_id" validate:"required"`
}

// GetProductUnitsWithConvertedPrices mengambil daftar unit yang tersedia untuk dibeli
// beserta harga pembelian yang sudah dikonversi
// Endpoint: GET /api/product-units (dengan body request)
func GetProductUnitsWithConvertedPrices(c *fiber.Ctx) error {
	db := configs.DB
	var req GetUnitsByProductIdRequest

	if err := c.BodyParser(&req); err != nil {
		// return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		// 	"message": "Invalid request body",
		// 	"error":   err.Error(),
		// })
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Body permintaan tidak valid", err)
	}

	if err := helpers.ValidateStruct(req); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Validasi product ID gagal", err)
	}

	branchID, _ := services.GetBranchID(c)
	if branchID == "" {
		return helpers.JSONResponse(c, fiber.StatusUnauthorized, "Branch ID tidak ditemukan di token. Unauthorized.", nil)
	}

	// 1. Dapatkan detail Product, khususnya Product.PurchasePrice (harga dasar) dan Product.UnitId (unit dasar)
	var product models.Product
	err := db.Where("id = ? AND branch_id = ?", req.ProductID, branchID).First(&product).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return helpers.JSONResponse(c, fiber.StatusNotFound, fmt.Sprintf("Product with ID %s not found in branch %s", req.ProductID, branchID), err)
		}
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil detail produk", err)
	}

	basePurchasePrice := product.PurchasePrice // Harga dasar (per unit dasar produk)
	baseUnitID := product.UnitId               // ID unit dasar produk

	// Map untuk menyimpan harga yang sudah dihitung untuk setiap unit ID
	// Ini akan menghindari duplikasi unit di respons akhir dan memudahkan lookup
	calculatedPrices := make(map[string]models.ProductUnitResponseItem)

	// Tambahkan unit dasar produk itu sendiri sebagai opsi pertama
	// Ini adalah harga acuan tanpa konversi
	calculatedPrices[baseUnitID] = models.ProductUnitResponseItem{
		UnitId:        baseUnitID,
		PurchasePrice: basePurchasePrice,
	}

	// 2. Dapatkan semua UnitConversion yang terkait dengan ProductID ini
	var unitConversions []models.UnitConversion
	err = db.Where("product_id = ? AND branch_id = ?", req.ProductID, branchID).Find(&unitConversions).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil daftar konversi satuan", err)
	}

	// Kumpulkan semua ID unit yang perlu dicari namanya (dari produk dasar, init_id, dan final_id konversi)
	unitIDsToFetch := []string{baseUnitID} // Mulai dengan unit dasar produk
	for _, uc := range unitConversions {
		unitIDsToFetch = append(unitIDsToFetch, uc.InitId)
		unitIDsToFetch = append(unitIDsToFetch, uc.FinalId)
	}

	// Dapatkan nama-nama unit yang relevan sekaligus
	unitNames := make(map[string]string)
	var units []models.Unit
	if len(unitIDsToFetch) > 0 {
		err = db.Where("id IN (?) AND branch_id = ?", unitIDsToFetch, branchID).Find(&units).Error
		if err != nil {
			return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil nama satuan", err)
		}
		for _, u := range units {
			unitNames[u.ID] = u.Name
		}
	}

	// Set nama untuk unit dasar
	if item, ok := calculatedPrices[baseUnitID]; ok {
		item.UnitName = unitNames[baseUnitID]
		calculatedPrices[baseUnitID] = item
	}

	// 3. Iterasi melalui UnitConversions dan hitung harga untuk setiap unit yang relevan
	// Kita harus mempertimbangkan konversi 'maju' (dari unit besar ke unit kecil)
	// dan konversi 'mundur' (dari unit kecil ke unit besar)
	// tergantung bagaimana ValueConv didefinisikan.

	// Untuk setiap konversi, tentukan harga untuk InitId dan FinalId
	for _, uc := range unitConversions {
		// Kasus 1: Konversi dari unit yang lebih besar ke unit yang lebih kecil (misal: Box ke Pcs, Strip ke Sachet)
		// Jika init_id adalah unit yang lebih besar, dan final_id adalah unit yang lebih kecil,
		// maka harga final_id = harga init_id / value_conv.
		// Asumsi: ValueConv selalu > 0.

		// Untuk menyederhanakan, kita akan selalu mengonversi ke/dari unit dasar (`baseUnitID`)
		// jika memungkinkan, untuk mendapatkan harga yang konsisten.

		// Skenario A: Konversi dari unit dasar ke unit lain (FinalId)
		// Contoh: 1 PCS (base) = 12 SACHET.
		// Product.PurchasePrice (harga per PCS) = 3000.
		// Harga Sachet = 3000 / 12 = 250.
		if uc.InitId == baseUnitID {
			if uc.ValueConv > 0 {
				priceForFinalID := basePurchasePrice / uc.ValueConv
				if existing, ok := calculatedPrices[uc.FinalId]; !ok || existing.PurchasePrice > priceForFinalID {
					// Hanya tambahkan/perbarui jika unit belum ada atau harga yang baru lebih murah
					// (Ini bisa jadi diperlukan jika ada beberapa jalur konversi, ambil yang paling optimal)
					calculatedPrices[uc.FinalId] = models.ProductUnitResponseItem{
						UnitId:        uc.FinalId,
						UnitName:      unitNames[uc.FinalId],
						PurchasePrice: priceForFinalID,
					}
				}
			}
		}

		// Skenario B: Konversi dari unit lain (InitId) ke unit dasar (FinalId)
		// Contoh: 1 BOX = 10 PCS (base).
		// Product.PurchasePrice (harga per PCS) = 250.
		// Harga Box = 250 * 10 = 2500.
		if uc.FinalId == baseUnitID {
			if uc.ValueConv > 0 {
				priceForInitID := basePurchasePrice * uc.ValueConv
				if existing, ok := calculatedPrices[uc.InitId]; !ok || existing.PurchasePrice > priceForInitID {
					// Hanya tambahkan/perbarui jika unit belum ada atau harga yang baru lebih murah
					calculatedPrices[uc.InitId] = models.ProductUnitResponseItem{
						UnitId:        uc.InitId,
						UnitName:      unitNames[uc.InitId],
						PurchasePrice: priceForInitID,
					}
				}
			}
		}
	}

	// 4. Ubah map menjadi slice untuk respons akhir
	var finalResponseItems []models.ProductUnitResponseItem
	for _, item := range calculatedPrices {
		if item.UnitName != "" { // Pastikan unit punya nama yang ditemukan
			finalResponseItems = append(finalResponseItems, item)
		}
	}

	// Opsional: Urutkan hasil jika diinginkan (misal berdasarkan nama unit atau harga)
	// sort.Slice(finalResponseItems, func(i, j int) bool {
	// 	return finalResponseItems[i].PurchasePrice < finalResponseItems[j].PurchasePrice
	// })

	// 5. Kembalikan respons sukses
	return helpers.JSONResponse(c, fiber.StatusOK, fmt.Sprintf("Units retrieved successfully for Product ID %s", req.ProductID), finalResponseItems)
}
