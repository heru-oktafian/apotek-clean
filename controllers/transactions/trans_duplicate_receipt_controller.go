package controllers

import (
	fmt "fmt"
	strings "strings"
	time "time"

	fiber "github.com/gofiber/fiber/v2"
	configs "github.com/heru-oktafian/fiber-apotek/configs"
	helpers "github.com/heru-oktafian/fiber-apotek/helpers"
	models "github.com/heru-oktafian/fiber-apotek/models"
	services "github.com/heru-oktafian/fiber-apotek/services"
	reports "github.com/heru-oktafian/fiber-apotek/services/reports"
	gorm "gorm.io/gorm"
)

// CreateDuplicateReceipt handles the creation of a new duplicate receipt record.
func CreateDuplicateReceipt(c *fiber.Ctx) error {
	// Hitung waktu saat ini di zona WIB
	nowWIB := time.Now().In(configs.Location)

	db := configs.DB
	var req DuplicateReceiptRequest

	// Deklarasi variabel 'err' untuk menangani error
	err := c.BodyParser(&req)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// Get default_member id dari token
	defaultMember, _ := services.GetClaimsToken(c, "default_member")

	//--- VALIDASI INPUT ---
	subscriptionType, _ := services.GetClaimsToken(c, "subscription_type")
	branchID, _ := services.GetBranchID(c)
	userID, _ := services.GetUserID(c)

	err = helpers.ValidateStruct(req)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Validate failed", err)
	}

	if req.DuplicateReceipt.Payment == "" {
		req.DuplicateReceipt.Payment = "paid_by_cash"
	}

	// --- Proses Penyimpanan Data ---
	// Mulai transaksi database
	tx := db.Begin()
	if tx.Error != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to begin database transaction", err)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Parse tanggal
	layout := "2006-01-02" // format harus YYYY-MM-DD
	parsedDate, err := time.Parse(layout, req.DuplicateReceipt.DuplicateReceiptDate)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD", err)
	}

	// 1. Simpan data Duplicate Receipt (induk)
	durID := helpers.GenerateID("DUR")

	// Inisilisasi total & profit duplicate receipt
	totalDUR := 0
	totalProfDUR := 0

	for i := range req.Items {
		itemID := helpers.GenerateID("DRI")
		req.Items[i].ID = itemID
		req.Items[i].DuplicateReceiptId = durID

		// Dapatkan detail produk dari database
		var product models.Product
		err = tx.Where("id = ?", req.Items[i].ProductId).First(&product).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return helpers.JSONResponse(c, fiber.StatusNotFound, "Product with ID %s not found", err)
			}
			return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to retrieve product details", err)
		}

		// Periksa ketersediaan stok produk
		if product.Stock < req.Items[i].Qty {
			tx.Rollback()
			return helpers.JSONResponse(c, fiber.StatusBadRequest, fmt.Sprintf("Insufficient stock for product %s. Available: %d, Requested: %d", product.Name, product.Stock, req.Items[i].Qty), err)
		}

		// Kurangi stok produk
		newStock := product.Stock - req.Items[i].Qty

		// Update stock in Redis

		err = tx.Model(&models.Product{}).Where("id = ?", product.ID).Update("stock", newStock).Error
		if err != nil {
			tx.Rollback()
			return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to update stock for product", err)
		}

		// Kalkulasi total_duplicate_recipe dan profit_estimate dari item-item
		totalDUR += req.Items[i].SubTotal
		// Profit per item = (Harga Jual - Harga Beli) * Qty
		totalProfDUR += (req.Items[i].Price - product.PurchasePrice) * req.Items[i].Qty
	}

	// Hitung total duplicate receipt dari item-item yang ada
	req.DuplicateReceipt.TotalDuplicateReceipt = totalDUR
	req.DuplicateReceipt.ProfitEstimate = totalProfDUR

	// Tentukan member ID yang digunakan (gunakan defaultMember jika req.DuplicateReceipt.MemberId kosong)
	memberIDToUse := req.DuplicateReceipt.MemberId
	if memberIDToUse == "" {
		memberIDToUse = defaultMember
	}

	insertDuplicateReceipt := models.DuplicateReceipts{
		ID:                    durID,
		MemberId:              memberIDToUse,
		Description:           req.DuplicateReceipt.Description,
		DuplicateReceiptDate:  parsedDate,
		TotalDuplicateReceipt: req.DuplicateReceipt.TotalDuplicateReceipt,
		ProfitEstimate:        req.DuplicateReceipt.ProfitEstimate,
		Payment:               req.DuplicateReceipt.Payment,
		BranchID:              branchID,
		UserID:                userID,
		CreatedAt:             nowWIB,
		UpdatedAt:             nowWIB,
	}

	err = tx.Create(&insertDuplicateReceipt).Error
	if err != nil {
		tx.Rollback()
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to create duplicate receipt", err)
	}

	err = tx.CreateInBatches(&req.Items, len(req.Items)).Error
	if err != nil {
		tx.Rollback()
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to create duplicate items", err)
	}

	transactionReportID := durID // Gunakan ID yang sama dengan DuplicateReceipt ID
	transactionReport := models.TransactionReports{
		ID:              transactionReportID,
		TransactionType: models.Sale, // Tipe transaksi adalah "sale"
		UserID:          userID,
		BranchID:        branchID,
		Total:           req.DuplicateReceipt.TotalDuplicateReceipt,
		Payment:         req.DuplicateReceipt.Payment,
		CreatedAt:       nowWIB,
		UpdatedAt:       nowWIB,
	}
	err = tx.Create(&transactionReport).Error
	if err != nil {
		tx.Rollback()
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to create transaction report", err)
	}

	// 4. Sinkronisasi laporan duplicate receipt
	// Inisialisasi laporan duplicate receipt
	var dailyProfit models.DailyProfitReport

	reportDate := parsedDate.Format("2006-01-02") // Format tanggal menjadi "YYYY-MM-DD"
	err = tx.Where("report_date = ? AND branch_id = ? AND user_id = ?", reportDate, branchID, userID).First(&dailyProfit).Error

	// Cek error selain record not found
	if err != nil && err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to check daily profit report", err)
	}

	if err == gorm.ErrRecordNotFound {
		// Jika belum ada, buat entri baru
		dailyProfitID := helpers.GenerateID("DPR")
		dailyProfit = models.DailyProfitReport{
			ID:             dailyProfitID,
			ReportDate:     parsedDate,
			UserID:         userID,
			BranchID:       branchID,
			TotalSales:     req.DuplicateReceipt.TotalDuplicateReceipt,
			ProfitEstimate: req.DuplicateReceipt.ProfitEstimate,
			CreatedAt:      nowWIB,
			UpdatedAt:      nowWIB,
		}
		err = tx.Create(&dailyProfit).Error
		if err != nil {
			tx.Rollback()
			return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to create daily profit report", err)
		}
	} else {
		// Jika sudah ada, update total_sales dan profit_estimate yang sudah ada
		dailyProfit.TotalSales += req.DuplicateReceipt.TotalDuplicateReceipt
		dailyProfit.ProfitEstimate += req.DuplicateReceipt.ProfitEstimate
		dailyProfit.UpdatedAt = time.Now()
		err = tx.Save(&dailyProfit).Error
		if err != nil {
			tx.Rollback()
			return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to update daily profit report", err)
		}
	}

	if subscriptionType == "quota" {
		var branch models.Branch
		err = tx.Where("id = ?", branchID).First(&branch).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return helpers.JSONResponse(c, fiber.StatusNotFound, "Branch not found", err)
			}
			return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to retrieve branch details for quota update", err)
		}

		if branch.Quota > 0 {
			branch.Quota -= 1
			err = tx.Save(&branch).Error
			if err != nil {
				tx.Rollback()
				return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to update quota for branch", err)
			}
		} else {
			tx.Rollback()
			return helpers.JSONResponse(c, fiber.StatusBadRequest, "No quota available for branch", nil)
		}
	}

	if req.DuplicateReceipt.MemberId != "" && req.DuplicateReceipt.MemberId != defaultMember {
		var member models.Member
		err = tx.Where("id = ?", req.DuplicateReceipt.MemberId).First(&member).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return helpers.JSONResponse(c, fiber.StatusNotFound, "Member not found", err)
			}
			return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to retrieve member details for points calculation", err)
		}

		var memberCategory models.MemberCategory
		err = tx.Where("id = ?", member.MemberCategoryId).First(&memberCategory).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return helpers.JSONResponse(c, fiber.StatusNotFound, "Member category not found", err)
			}
			return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to retrieve member category for points calculation", err)
		}

		if memberCategory.PointsConversionRate > 0 {
			// Pastikan total_sale adalah float untuk perhitungan poin
			pointsEarned := float64(req.DuplicateReceipt.TotalDuplicateReceipt) / float64(memberCategory.PointsConversionRate)
			member.Points += int(pointsEarned) // Tambahkan poin yang didapat (gunakan int jika kolom points int)

			err = tx.Save(&member).Error
			if err != nil {
				tx.Rollback()
				return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to update points for member", err)
			}
		} else {
			// Optional: Handle case where PointsConversionRate is 0 or less
			// You might want to log this or return a specific error
			fmt.Printf("Warning: PointsConversionRate for member category %d is zero or negative. Points not calculated.\n", member.MemberCategoryId)
		}
	}

	// Commit transaksi jika semua berhasil
	err = tx.Commit().Error
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to commit database transaction", err)
	}

	// --- FORMAT DATA RESPONS ---
	// Buat itemsResponse tanpa field duplicate_receipt_id agar sesuai permintaan user
	var itemsResponse []fiber.Map
	for _, item := range req.Items {
		itemsResponse = append(itemsResponse, fiber.Map{
			"id":         item.ID,
			"product_id": item.ProductId,
			"price":      item.Price,
			"qty":        item.Qty,
			"sub_total":  item.SubTotal,
		})
	}

	// Bungkus data hasil transaksi ke dalam format respons yang diinginkan
	dataResponse := fiber.Map{
		"duplicate_receipt": fiber.Map{
			"id":                      insertDuplicateReceipt.ID,
			"member_id":               insertDuplicateReceipt.MemberId,
			"description":             insertDuplicateReceipt.Description,
			"duplicate_receipt_date":  insertDuplicateReceipt.DuplicateReceiptDate.Format("2006-01-02"),
			"total_duplicate_receipt": insertDuplicateReceipt.TotalDuplicateReceipt,
			"profit_estimate":         insertDuplicateReceipt.ProfitEstimate,
			"payment":                 insertDuplicateReceipt.Payment,
		},
		"items": itemsResponse,
	}

	// Berhasil
	return helpers.JSONResponse(c, fiber.StatusOK, "Duplicate receipt transaction created successfully", dataResponse)
}

// UpdateDuplicateReceipt Function (Modified)
func UpdateDuplicateReceipt(c *fiber.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	branchID, _ := services.GetBranchID(c)
	userID, _ := services.GetUserID(c) // Get default_member id dari token
	defaultMember, _ := services.GetClaimsToken(c, "default_member")

	total_before := 0
	profit_before := 0

	db := configs.DB
	id := c.Params("id")

	var duplicate_receipt models.DuplicateReceipts
	if err := db.First(&duplicate_receipt, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Receipt not found", err)
	}

	total_before += duplicate_receipt.TotalDuplicateReceipt
	profit_before += duplicate_receipt.ProfitEstimate

	var input models.DuplicateReceiptInput
	if err := c.BodyParser(&input); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid input", err)
	}

	if input.MemberId != "" {
		var member models.Member
		if err := db.Where("id = ?", input.MemberId).First(&member).Error; err != nil {
			// Jika ID tidak valid, fallback ke default
			memberId, _ := services.GetClaimsToken(c, "default_member")
			duplicate_receipt.MemberId = memberId
		} else {
			duplicate_receipt.MemberId = defaultMember
		}
	}
	// Jika kosong → tidak diubah, tetap pakai MemberID yang sudah ada

	if input.Payment != "" {
		duplicate_receipt.Payment = models.PaymentStatus(input.Payment)
	}

	duplicate_receipt.UpdatedAt = nowWIB

	var items []models.DuplicateReceiptItems
	if err := db.Where("duplicate_receipt_id = ?", id).Find(&items).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to fetch sale items", err)
	}

	total := 0
	for _, item := range items {
		total += item.SubTotal
	}

	profit := 0
	for _, item := range items {
		profit += item.SubTotal - item.Price*item.Qty
	}

	if err := db.Save(&duplicate_receipt).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to update Duplicate receipt", err)
	}

	if err := reports.SyncDuplicateReceiptReport(db, duplicate_receipt); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to sync Duplicate receipt report", err)
	}

	// Sync laporan penjualan agar tetap konsisten
	_ = reports.SyncDailyProfitReport(db, branchID, userID, duplicate_receipt.DuplicateReceiptDate, total, profit, total_before, profit_before)

	return helpers.JSONResponse(c, fiber.StatusOK, "Duplicate receipt updated successfully", duplicate_receipt)
}

// DeleteDuplicateReceipt Function
func DeleteDuplicateReceipt(c *fiber.Ctx) error {
	db := configs.DB
	id := c.Params("id")

	// Ambil duplicate receipt
	var duplicate_receipt models.DuplicateReceipts
	if err := db.First(&duplicate_receipt, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Duplicate receipt not found", err)
	}

	// Ambil & hapus item, serta rollback stok
	var items []models.DuplicateReceiptItems
	if err := db.Where("duplicate_receipt_id = ?", id).Find(&items).Error; err == nil {
		for _, item := range items {
			_ = services.SubtractProductStock(db, item.ProductId, item.Qty)
		}
		db.Where("duplicate_receipt_id = ?", id).Delete(&models.DuplicateReceiptItems{})
	}

	// Hapus laporan transaksi
	if err := db.Where("id = ? AND transaction_type = ?", duplicate_receipt.ID, models.Sale).Delete(&models.TransactionReports{}).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to delete transaction report", err)
	}

	// Hapus data penjualan
	if err := db.Delete(&duplicate_receipt).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to delete sale", err)
	}

	// Delete laporan profit harian asynchronously
	go func() {
		if err := reports.DeleteDailyProfitReport(db, id, "duplicate_receipt"); err != nil {
			fmt.Printf("Failed to delete daily profit report asynchronously: %v\n", err)
		}
	}()

	return helpers.JSONResponse(c, fiber.StatusOK, "Duplicate receipt deleted successfully", duplicate_receipt)
}

type DuplicateReceiptRequest struct {
	DuplicateReceipt models.DuplicateReceiptInput   `json:"duplicate_receipt"`
	Items            []models.DuplicateReceiptItems `json:"items"`
}

// CreateDuplicateReceiptItem Function
func CreateDuplicateReceiptItem(c *fiber.Ctx) error {
	// Get branch and user IDs from middleware
	var item models.DuplicateReceiptItems

	db := configs.DB

	if err := c.BodyParser(&item); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid input", err)
	}

	// Ambil harga jual produk dari tabel products
	var product models.Product
	if err := db.Select("sales_price").Where("id = ?", item.ProductId).First(&product).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to fetch product price", err)
	}

	// Gunakan sales_price dari produk, abaikan inputan frontend
	item.Price = product.SalesPrice

	// Cek apakah item dengan duplicate_receipt_id dan product_id sudah ada
	var existing models.DuplicateReceiptItems
	err := db.Where("duplicate_receipt_id = ? AND product_id = ?", item.DuplicateReceiptId, item.ProductId).First(&existing).Error
	if err == nil {
		// Sudah ada: update qty dan sub_total
		existing.Qty += item.Qty
		existing.Price = product.SalesPrice
		existing.SubTotal = existing.Qty * existing.Price

		if err := db.Save(&existing).Error; err != nil {
			return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to update sale item", err)
		}

		if err := services.ReduceProductStock(db, item.ProductId, item.Qty); err != nil {
			return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to reduce product stock", err)
		}

		// Supporting operations asynchronously
		go func() {
			// Update stock in Redis
			if err := reports.RecalculateTotalDuplicate(db, item.DuplicateReceiptId); err != nil {
				fmt.Printf("Failed to recalculate total duplicate asynchronously: %v\n", err)
			}
		}()

		return helpers.JSONResponse(c, fiber.StatusOK, "Item updated successfully", existing)

	} else if err != gorm.ErrRecordNotFound {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to find existing sale item", err)
	}

	// Data belum ada, buat item baru
	if item.ID == "" {
		item.ID = helpers.GenerateID("DRI")
	}
	item.SubTotal = item.Qty * item.Price

	if err := db.Create(&item).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to create sale item", err)
	}

	if err := services.ReduceProductStock(db, item.ProductId, item.Qty); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to reduce product stock", err)
	}

	// Recalculate total and sync reports asynchronously
	go func() {
		// Update stock in Redis
		if err := reports.RecalculateTotalDuplicate(db, item.DuplicateReceiptId); err != nil {
			fmt.Printf("Failed to recalculate total duplicate asynchronously: %v\n", err)
		}

		// Sync laporan profit harian
		var duplicateReceipt models.DuplicateReceipts
		if err := db.First(&duplicateReceipt, "id = ?", item.DuplicateReceiptId).Error; err != nil {
			fmt.Printf("Failed to fetch duplicate receipt asynchronously: %v\n", err)
			return
		}

		if err := reports.SyncDailyProfitReport(db, duplicateReceipt.BranchID, duplicateReceipt.UserID, duplicateReceipt.DuplicateReceiptDate, duplicateReceipt.TotalDuplicateReceipt, duplicateReceipt.ProfitEstimate, 0, 0); err != nil {
			fmt.Printf("Failed to sync daily profit report asynchronously: %v\n", err)
		}
	}()

	return helpers.JSONResponse(c, fiber.StatusOK, "Item added successfully", item)
}

// UpdateDuplicateReceiptItem Function
func UpdateDuplicateReceiptItem(c *fiber.Ctx) error {
	db := configs.DB
	id := c.Params("id")

	var existingItem models.DuplicateReceiptItems
	if err := db.First(&existingItem, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Item not found", err)
	}

	// Parsing data baru dari body (hanya untuk ambil ProductId dan Qty baru)
	var updatedData struct {
		ProductId string `json:"product_id"`
		Qty       int    `json:"qty"`
	}
	if err := c.BodyParser(&updatedData); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid input", err)
	}

	// Rollback stok lama
	if err := services.AddProductStock(db, existingItem.ProductId, existingItem.Qty); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to add product stock", err)
	}

	// Ambil harga jual dari produk baru
	var product models.Product
	if err := db.Select("sales_price").Where("id = ?", updatedData.ProductId).First(&product).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to get product price", err)
	}

	// Kurangi stok baru
	if err := services.ReduceProductStock(db, updatedData.ProductId, updatedData.Qty); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to reduce product stock", err)
	}

	// Update item
	existingItem.ProductId = updatedData.ProductId
	existingItem.Qty = updatedData.Qty
	existingItem.Price = product.SalesPrice
	existingItem.SubTotal = product.SalesPrice * updatedData.Qty

	if err := db.Save(&existingItem).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to update sale item", err)
	}

	// Supporting operations asynchronously
	go func() {
		if err := reports.RecalculateTotalDuplicate(db, existingItem.DuplicateReceiptId); err != nil {
			fmt.Printf("Failed to recalculate total duplicate asynchronously: %v\n", err)
		}

		// Sync laporan profit harian
		var duplicateReceipt models.DuplicateReceipts
		if err := db.First(&duplicateReceipt, "id = ?", existingItem.DuplicateReceiptId).Error; err != nil {
			fmt.Printf("Failed to fetch duplicate receipt asynchronously: %v\n", err)
			return
		}

		if err := reports.SyncDailyProfitReport(db, duplicateReceipt.BranchID, duplicateReceipt.UserID, duplicateReceipt.DuplicateReceiptDate, duplicateReceipt.TotalDuplicateReceipt, duplicateReceipt.ProfitEstimate, 0, 0); err != nil {
			fmt.Printf("Failed to sync daily profit report asynchronously: %v\n", err)
		}
	}()

	return helpers.JSONResponse(c, fiber.StatusOK, "Item updated successfully", existingItem)
}

// Delete DuplicateReceiptItem
func DeleteDuplicateReceiptItem(c *fiber.Ctx) error {
	db := configs.DB
	id := c.Params("id")

	var item models.DuplicateReceiptItems
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Item not found", err)
	}

	// Rollback stok
	if err := services.AddProductStock(db, item.ProductId, item.Qty); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to add product stock", err)
	}

	// Hapus item
	if err := db.Delete(&item).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to delete sale item", err)
	}

	// Supporting operations asynchronously
	go func() {
		// Update stock in Redis
		if err := reports.RecalculateTotalDuplicate(db, item.DuplicateReceiptId); err != nil {
			fmt.Printf("Failed to recalculate total duplicate asynchronously: %v\n", err)
		}
	}()

	return helpers.JSONResponse(c, fiber.StatusOK, "Item deleted successfully", item)
}

// GetAllDuplicateReceipts tampilkan semua duplicate receipt items
func GetAllDuplicateReceipts(c *fiber.Ctx) error {
	// Ambil Branch ID dari user yang login
	branchID, _ := services.GetBranchID(c)

	var salesFromDB []models.AllDuplicateReceipts // Gunakan models.AllDuplicateReceipts untuk mengambil data dari DB

	query := configs.DB.Table("duplicate_receipts dr").
		Select("dr.id, dr.member_id, mbr.name AS member_name, dr.duplicate_receipt_date, dr.total_duplicate_receipt, dr.profit_estimate, dr.payment").
		Joins("LEFT JOIN members mbr on mbr.id = dr.member_id").
		Where("dr.branch_id = ? AND dr.total_duplicate_receipt > 0", branchID).
		Order("dr.created_at DESC")

	// panggil helper paginate dengan parameter search dan month
	_, search, total, page, totalPages, err := helpers.PaginateWithSearchAndMonth(
		c,
		query,
		&salesFromDB,
		[]string{"mbr.name"},
		"dr.duplicate_receipt_date",
		1,
		10,
	)

	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil duplicate receipt", err.Error())
	}

	// Buat slice baru untuk menampung data yang sudah diformat
	var formattedDuplicateData []models.DuplicateDetailResponse
	for _, duplicate_receipt := range salesFromDB {
		formattedDuplicateData = append(formattedDuplicateData, models.DuplicateDetailResponse{
			ID:                    duplicate_receipt.ID,
			MemberId:              duplicate_receipt.MemberId,
			MemberName:            duplicate_receipt.MemberName,
			DuplicateReceiptDate:  helpers.FormatIndonesianDate(duplicate_receipt.DuplicateReceiptDate), // Format tanggal di sini
			TotalDuplicateReceipt: duplicate_receipt.TotalDuplicateReceipt,
			ProfitEstimate:        duplicate_receipt.ProfitEstimate,
			Payment:               string(duplicate_receipt.Payment),
		})
	}

	// Gunakan JSONResponseGetAll helper dengan data yang sudah diformat
	return helpers.JSONResponseGetAll(
		c,
		fiber.StatusOK,
		"Duplicate receipts retrieved successfully",
		search,
		int(total),
		page,
		totalPages,
		10,
		formattedDuplicateData, // Kirim data yang sudah diformat (slice dari DuplicateDetailResponse)
	)
}

// GetAllDuplicateDetail menampilkan sales dengan kolom description yang
// berisi daftar nama item (dipisah koma) diikuti dengan duplicate_receipt_date (DD-MM-YYYY HH:MM)
// di mana waktu ditambah 7 jam sesuai permintaan.
func GetAllDuplicateDetail(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)

	// Struct sementara untuk menampung hasil query
	type duplicateSummary struct {
		ID                    string
		TotalDuplicateReceipt int
		Description           string
		DuplicateReceiptDate  time.Time
		Payment               string
		UpdatedAt             time.Time
	}

	var salesFromDB []duplicateSummary

	query := configs.DB.Table("duplicate_receipts dr").
		Select("dr.id, dr.total_duplicate_receipt, dr.payment, dr.duplicate_receipt_date, dr.description, dr.updated_at").
		Joins("LEFT JOIN members mbr on mbr.id = dr.member_id").
		Where("dr.branch_id = ? AND dr.total_duplicate_receipt > 0", branchID).
		Order("dr.created_at DESC")

	// panggil helper paginate dengan parameter search dan month
	_, search, total, page, totalPages, err := helpers.PaginateWithSearchAndMonth(
		c,
		query,
		&salesFromDB,
		[]string{"mbr.name"},
		"dr.duplicate_receipt_date",
		1,
		10,
	)

	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil duplicate receipt", err.Error())
	}

	// Bentuk response dengan description
	var formatted []map[string]interface{}
	for _, s := range salesFromDB {
		// Ambil nama item untuk sale ini
		var itemNames []string
		if err := configs.DB.Table("duplicate_receipt_items dri").
			Select("pro.name").
			Joins("LEFT JOIN products pro ON pro.id = dri.product_id").
			Where("dri.duplicate_receipt_id = ?", s.ID).
			Order("pro.name ASC").
			Pluck("pro.name", &itemNames).Error; err != nil {
			return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to get duplicate receipt items", err)
		}

		// Gabungkan nama item, lalu tambahkan tanggal yang ditambah 7 jam
		descItems := strings.Join(itemNames, ", ")
		dateWith7 := s.UpdatedAt.Add(7 * time.Hour).Format("02-01-2006 15:04")
		var description string
		if descItems != "" {
			description = s.Description + " ; " + descItems + " ; " + dateWith7
		} else {
			description = s.Description + " ; " + dateWith7
		}

		formatted = append(formatted, map[string]interface{}{
			"id":                      s.ID,
			"duplicate_receipt_date":  helpers.FormatIndonesianDate(s.DuplicateReceiptDate),
			"description":             description,
			"payment":                 s.Payment,
			"total_duplicate_receipt": s.TotalDuplicateReceipt,
		})
	}

	return helpers.JSONResponseGetAll(
		c,
		fiber.StatusOK,
		"Duplicate receipts retrieved successfully",
		search,
		int(total),
		page,
		totalPages,
		10,
		formatted,
	)
}

// GetAllDuplicateItems tampilkan semua item berdasarkan duplicate_receipt_id tanpa pagination
func GetAllDuplicateItems(c *fiber.Ctx) error {
	// Get duplicate receipt id dari param
	duplicateReceiptID := c.Params("id")

	// Parsing body JSON ke struct
	var DuplicateItems []models.AllDuplicateReceiptItems

	// Query dasar
	query := configs.DB.Table("duplicate_receipt_items dri").
		Select("dri.id, dri.duplicate_receipt_id, dri.product_id, pro.name AS product_name, dri.price, dri.qty, un.name AS unit_name, dri.sub_total").
		Joins("LEFT JOIN products pro ON pro.id = dri.product_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("dri.duplicate_receipt_id = ?", duplicateReceiptID).
		Order("pro.name ASC")

	// Eksekusi query
	if err := query.Scan(&DuplicateItems).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get items failed", err)
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Items retrieved successfully", DuplicateItems)
}

// GetDuplicateWithItems menampilkan satu duplicate receipt beserta semua item-nya
func GetDuplicateWithItems(c *fiber.Ctx) error {
	db := configs.DB

	duplicateId := c.Params("id")

	// Gunakan models.AllDuplicateReceipts untuk mengambil data dari DB
	var duplicate_receipts models.AllDuplicateReceipts

	err := db.Table("duplicate_receipts dr").
		Select("dr.id, dr.member_id, mbr.name AS member_name, dr.duplicate_receipt_date, dr.total_duplicate_receipt, dr.profit_estimate, dr.payment").
		Joins("LEFT JOIN members mbr ON mbr.id = dr.member_id").
		Where("dr.id = ?", duplicateId).
		Scan(&duplicate_receipts).Error
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to get duplicate receipt", err)
	}

	// Ambil item pembelian terkait
	var items []models.AllDuplicateReceiptItems
	err = db.Table("duplicate_receipt_items dri").
		Select("dri.id, dri.duplicate_receipt_id, dri.product_id, pro.name AS product_name, dri.price, dri.qty, un.name AS unit_name, dri.sub_total").
		Joins("LEFT JOIN products pro ON pro.id = dri.product_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("dri.duplicate_receipt_id = ?", duplicateId).
		Order("pro.name ASC").
		Scan(&items).Error

	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to get duplicate receipt items", err)
	}

	// Format tanggal secara manual untuk respons ini
	// Menggunakan helper FormatIndonesianDate yang sudah kita buat
	formattedDuplicateDate := helpers.FormatIndonesianDate(duplicate_receipts.DuplicateReceiptDate)

	// Buat objek respons menggunakan struct DuplicateItemResponse yang baru
	// dan isi field-fieldnya
	responseDetail := models.DuplicateItemResponse{
		ID:                    duplicate_receipts.ID,
		MemberId:              duplicate_receipts.MemberId,
		MemberName:            duplicate_receipts.MemberName,
		DuplicateReceiptDate:  formattedDuplicateDate, // Gunakan tanggal yang sudah diformat
		TotalDuplicateReceipt: duplicate_receipts.TotalDuplicateReceipt,
		ProfitEstimate:        duplicate_receipts.ProfitEstimate,
		Payment:               string(duplicate_receipts.Payment),
		Items:                 items,
	}

	// Panggil JSONResponse yang sudah ada, meneruskan DuplicateItemResponse sebagai 'data'
	return helpers.JSONResponse(c, fiber.StatusOK, "Duplicate receipt retrieved successfully", responseDetail)
}
