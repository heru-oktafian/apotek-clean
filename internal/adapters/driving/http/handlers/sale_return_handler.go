package handlers

import (
	fmt "fmt"
	strings "strings"
	time "time"

	configs "apotek-clean/configs"
	helpers "apotek-clean/helpers"
	models "apotek-clean/models"
	services "apotek-clean/services"
	fiber "github.com/gofiber/fiber/v2"
	gorm "gorm.io/gorm"
)

func rollbackSaleReturnWithJSON(c *fiber.Ctx, tx *gorm.DB, status int, message string, err interface{}) error {
	tx.Rollback()
	return helpers.JSONResponse(c, status, message, err)
}

func createSaleReturnTransactionReport(tx *gorm.DB, saleReturn models.SaleReturns, userID, branchID string, nowWIB time.Time) error {
	transactionReport := models.TransactionReports{
		ID:              helpers.GenerateID("TRX"),
		TransactionType: models.SaleReturn,
		UserID:          userID,
		BranchID:        branchID,
		Total:           saleReturn.TotalReturn,
		Payment:         saleReturn.Payment,
		CreatedAt:       nowWIB,
		UpdatedAt:       nowWIB,
	}
	return tx.Create(&transactionReport).Error
}

func applySaleReturnQuotaIfNeeded(tx *gorm.DB, subscriptionType, branchID string) error {
	if subscriptionType != "quota" {
		return nil
	}

	var branch models.Branch
	err := tx.Where("id = ?", branchID).First(&branch).Error
	if err != nil {
		return err
	}

	if branch.Quota <= 0 {
		return fmt.Errorf("quota exhausted")
	}

	branch.Quota -= 1
	return tx.Save(&branch).Error
}

// CreateSaleReturnTransaction adalah fungsi untuk membuat transaksi retur penjualan baru
func CreateSaleReturnTransaction(c *fiber.Ctx) error {
	nowWIB := time.Now().In(configs.Location)

	subscriptionType, _ := services.GetClaimsToken(c, "subscription_type")
	branchID, _ := services.GetClaimsToken(c, "branch_id")
	userID, _ := services.GetClaimsToken(c, "user_id")

	db := configs.DB
	var req models.SaleReturnRequest // pastikan struct request ini ada
	err := c.BodyParser(&req)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Body permintaan tidak valid", err.Error())
	}

	if req.SaleReturn.Payment == "" {
		req.SaleReturn.Payment = "paid_by_cash"
	}

	req.SaleReturn.UserID = userID
	req.SaleReturn.BranchID = branchID

	err = helpers.ValidateStruct(req.SaleReturn)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Validasi input retur penjualan gagal", err.Error())
	}

	for _, item := range req.SaleReturnItems {
		err = helpers.ValidateStruct(item)
		if err != nil {
			return helpers.JSONResponse(c, fiber.StatusBadRequest, "Validasi salah satu item retur penjualan gagal", err.Error())
		}
	}

	returnDate, err := services.ParseSaleReturnDate(req.SaleReturn.ReturnDate, nowWIB)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Format return_date tidak valid. Gunakan YYYY-MM-DD.", err.Error())
	}

	tx := db.Begin()
	if tx.Error != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal memulai transaksi database", tx.Error.Error())
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Validasi apakah sale_id valid
	var sale models.Sales
	err = tx.Where("id = ?", req.SaleReturn.SaleId).First(&sale).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return rollbackSaleReturnWithJSON(c, tx, fiber.StatusNotFound, fmt.Sprintf("Penjualan dengan ID %s tidak ditemukan", req.SaleReturn.SaleId), err.Error())
		}
		return rollbackSaleReturnWithJSON(c, tx, fiber.StatusInternalServerError, "Gagal mengambil data penjualan", err.Error())
	}

	saleReturnID := helpers.GenerateID("SRT")
	saleReturn := models.SaleReturns{
		ID:         saleReturnID,
		SaleId:     req.SaleReturn.SaleId,
		ReturnDate: returnDate,
		BranchID:   branchID,
		Payment:    req.SaleReturn.Payment,
		UserID:     userID,
		CreatedAt:  nowWIB,
		UpdatedAt:  nowWIB,
	}

	var totalReturn int
	var saleReturnItems []models.SaleReturnItems

	for _, item := range req.SaleReturnItems {
		parsedExpiredDate, err := time.Parse("2006-01-02", item.ExpiredDate)
		if err != nil {
			return rollbackSaleReturnWithJSON(c, tx, fiber.StatusBadRequest, fmt.Sprintf("expired_date tidak valid untuk produk %s", item.ProductId), err.Error())
		}

		lookup, err := services.LookupSaleReturnSaleItem(tx, req.SaleReturn.SaleId, item.ProductId)
		if err != nil {
			return rollbackSaleReturnWithJSON(c, tx, fiber.StatusBadRequest, fmt.Sprintf("Produk %s tidak ditemukan pada penjualan asal", item.ProductId), err.Error())
		}

		totalReturnedQty, err := services.LookupSaleReturnReturnedQty(tx, req.SaleReturn.SaleId, item.ProductId)
		if err != nil {
			return rollbackSaleReturnWithJSON(c, tx, fiber.StatusInternalServerError, "Gagal memeriksa retur sebelumnya", err.Error())
		}
		saleItem := lookup.SaleItem

		if err := services.ValidateSaleReturnQuantity(saleItem.Qty, item.Qty, totalReturnedQty, item.ProductId); err != nil {
			return rollbackSaleReturnWithJSON(c, tx, fiber.StatusBadRequest, fmt.Sprintf("Total qty retur untuk produk %s melebihi jumlah yang dijual. Dijual: %d, Sudah Diretur: %d, Retur Ini: %d",
				item.ProductId, saleItem.Qty, totalReturnedQty, item.Qty), nil)
		}

		// Ambil informasi produk
		var product models.Product
		err = tx.Where("id = ?", item.ProductId).First(&product).Error
		if err != nil {
			return rollbackSaleReturnWithJSON(c, tx, fiber.StatusInternalServerError, fmt.Sprintf("Gagal mengambil info produk untuk %s", item.ProductId), err.Error())
		}

		preparedItem := services.PrepareSaleReturnItem(helpers.GenerateID("SRI"), saleReturnID, item, saleItem.Price, parsedExpiredDate)

		// Update stok
		err = tx.Model(&models.Product{}).Where("id = ?", item.ProductId).
			Update("stock", gorm.Expr("stock + ?", preparedItem.ActualQtyToAdd)).Error
		if err != nil {
			return rollbackSaleReturnWithJSON(c, tx, fiber.StatusInternalServerError, fmt.Sprintf("Gagal memperbarui stok untuk produk %s", item.ProductId), err.Error())
		}

		totalReturn += preparedItem.SubTotal
		saleReturnItems = append(saleReturnItems, preparedItem.SaleReturnItem)

	}

	saleReturn.TotalReturn = totalReturn

	err = tx.Create(&saleReturn).Error
	if err != nil {
		return rollbackSaleReturnWithJSON(c, tx, fiber.StatusInternalServerError, "Gagal membuat retur penjualan", err.Error())
	}

	err = tx.CreateInBatches(&saleReturnItems, len(saleReturnItems)).Error
	if err != nil {
		return rollbackSaleReturnWithJSON(c, tx, fiber.StatusInternalServerError, "Gagal membuat item retur penjualan", err.Error())
	}

	err = createSaleReturnTransactionReport(tx, saleReturn, userID, branchID, nowWIB)
	if err != nil {
		return rollbackSaleReturnWithJSON(c, tx, fiber.StatusInternalServerError, "Gagal membuat laporan transaksi retur penjualan", err.Error())
	}

	err = applySaleReturnQuotaIfNeeded(tx, subscriptionType, branchID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return rollbackSaleReturnWithJSON(c, tx, fiber.StatusInternalServerError, "Gagal mengambil info cabang untuk kuota", err.Error())
		}
		if err.Error() == "quota exhausted" {
			return rollbackSaleReturnWithJSON(c, tx, fiber.StatusBadRequest, "Kuota cabang sudah habis", nil)
		}
		return rollbackSaleReturnWithJSON(c, tx, fiber.StatusInternalServerError, "Gagal memperbarui kuota cabang", err.Error())
	}

	err = tx.Commit().Error
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal melakukan commit transaksi", err.Error())
	}

	response := services.BuildSaleReturnResponse(saleReturn, saleReturnItems)
	response["return_date"] = helpers.FormatIndonesianDate(saleReturn.ReturnDate)
	return helpers.JSONResponse(c, fiber.StatusOK, "Transaksi retur penjualan berhasil dibuat", response)
}

// GetSaleItemsForReturn digunakan untuk mengambil item penjualan yang bisa diretur
func GetSaleItemsForReturn(c *fiber.Ctx) error {
	saleId := c.Query("sale_id")
	if saleId == "" {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "sale_id wajib diisi", nil)
	}

	var results []struct {
		ProID    string `json:"pro_id"`
		ProName  string `json:"pro_name"`
		Stock    int    `json:"stock"`
		UnitID   string `json:"unit_id"`
		UnitName string `json:"unit_name"`
		Price    int    `json:"price"`
	}

	err := configs.DB.Raw(`
        SELECT 
            A.product_id AS pro_id,
            B.name AS pro_name,
            A.qty AS stock,
            B.unit_id,
            C.name AS unit_name,
            A.price
        FROM sale_items A
        LEFT JOIN products B ON B.id = A.product_id
        LEFT JOIN units C ON C.id = B.unit_id
        LEFT JOIN (
            SELECT 
                sri.product_id, 
                SUM(sri.qty) AS total_returned
            FROM sale_return_items sri
            INNER JOIN sale_returns sr ON sri.sale_return_id = sr.id
            WHERE sr.sale_id = ?
            GROUP BY sri.product_id
        ) R ON R.product_id = A.product_id
        WHERE A.sale_id = ? 
        AND COALESCE(R.total_returned, 0) < A.qty
    `, saleId, saleId).Scan(&results).Error

	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil item penjualan", err.Error())
	}

	if len(results) == 0 {
		return helpers.JSONResponse(c, fiber.StatusOK, "Tidak ada item yang bisa diretur untuk penjualan ini", []struct {
			ProID    string `json:"pro_id"`
			ProName  string `json:"pro_name"`
			Stock    int    `json:"stock"`
			UnitID   string `json:"unit_id"`
			UnitName string `json:"unit_name"`
			Price    int    `json:"price"`
		}{})
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Data item retur ditemukan", results)
}

// GetSaleReturnWithItems menampilkan satu retur penjualan beserta semua item-nya
func GetSaleReturnWithItems(c *fiber.Ctx) error {
	db := configs.DB

	saleReturnID := c.Params("id")

	// Gunakan models.AllSaleReturns untuk mengambil data dari DB
	var saleReturn models.AllSaleReturns

	err := db.Table("sale_returns A").
		Select("A.id, A.sale_id, A.return_date, A.payment, A.total_return").
		Where("A.id = ?", saleReturnID).
		Scan(&saleReturn).Error

	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data retur penjualan", err.Error())
	}

	// Ambil item retur penjualan terkait
	var items []models.AllSaleReturnItems
	err = db.Table("sale_return_items A").
		Select("A.id, A.sale_return_id, A.product_id AS pro_id, B.name AS pro_name, B.unit_id, C.name AS unit_name, A.qty, A.price, A.sub_total, A.expired_date").
		Joins("LEFT JOIN products B on B.id=A.product_id").
		Joins("LEFT JOIN units C on C.id=B.unit_id").
		Where("A.sale_return_id = ?", saleReturnID).
		Order("B.name ASC").
		Scan(&items).Error

	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil item retur penjualan", err.Error())
	}

	// Format tanggal secara manual untuk respons ini
	formattedSaleReturnDate := helpers.FormatIndonesianDate(saleReturn.ReturnDate)

	// Buat objek respons menggunakan struct SaleItemResponse yang baru
	responseDetail := models.SaleReturnItemResponse{
		ID:          saleReturn.ID,
		SaleId:      saleReturn.SaleId,
		ReturnDate:  formattedSaleReturnDate,
		TotalReturn: saleReturn.TotalReturn,
		Payment:     string(saleReturn.Payment),
		Items:       items,
	}

	// Panggil JSONResponse yang sudah ada, meneruskan SaleItemResponse sebagai 'data'
	return helpers.JSONResponse(c, fiber.StatusOK, "Retur penjualan berhasil diambil", responseDetail)
}

// GetAllSaleReturns menampilkan semua retur penjualan
func GetAllSaleReturns(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)

	var saleReturnsFromDB []models.AllSaleReturns // Gunakan models.AllSaleReturns untuk mengambil data dari DB

	query := configs.DB.Table("sale_returns A").
		Select("A.id, A.sale_id, A.return_date, A.payment, A.total_return").
		Where("A.branch_id = ? ", branchID).
		Order("A.created_at DESC")

	// Panggil helper PaginateWithSearchAndMonth
	result, search, total, page, totalPages, err := helpers.PaginateWithSearchAndMonth(
		c,
		query,
		&saleReturnsFromDB,
		[]string{"A.sale_id"},
		"A.return_date",
		1,
		10,
	)

	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil retur penjualan", err)
	}

	saleReturnsData := *result.(*[]models.AllSaleReturns)

	// Buat slice baru untuk menampung data yang sudah diformat
	var formattedSaleReturnsData []models.SaleReturnsResponse
	for _, saleReturn := range saleReturnsData {
		formattedSaleReturnsData = append(formattedSaleReturnsData, models.SaleReturnsResponse{
			ID:          saleReturn.ID,
			SaleId:      saleReturn.SaleId,
			ReturnDate:  helpers.FormatIndonesianDate(saleReturn.ReturnDate), // Format tanggal di sini
			TotalReturn: saleReturn.TotalReturn,
			Payment:     string(saleReturn.Payment),
		})
	}

	// Gunakan JSONResponseGetAll helper dengan data yang sudah diformat
	return helpers.JSONResponseGetAll(
		c,
		fiber.StatusOK,
		"Data retur pembelian berhasil diambil",
		search,
		total,
		page,
		totalPages,
		10,
		formattedSaleReturnsData, // Kirim data yang sudah diformat (slice dari SaleReturnsResponse)
	)
}

// CmbSale mengambil data penjualan
func CmbSale(c *fiber.Ctx) error {
	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	branchID, _ := services.GetBranchID(c)

	// Ambil parameter search dan month dari query URL
	search := strings.TrimSpace(c.Query("search"))
	month := strings.TrimSpace(c.Query("month"))

	// Jika month kosong, isi dengan bulan ini (format YYYY-MM)
	if month == "" {
		month = nowWIB.Format("2006-01")
	}

	// Gunakan struct ringan hanya dengan kolom yang dibutuhkan untuk mengurangi I/O dan marshal
	var sales []struct {
		ID        string    `json:"id"`
		SaleDate  time.Time `json:"sale_date"`
		TotalSale int       `json:"total_sale"`
		MemberID  string    `json:"member_id"`
	}

	query := configs.DB.Table("sales").
		Select("id, sale_date, total_sale, member_id").
		Where("branch_id = ?", branchID)

	// Filter by month (sale_date)
	if month != "" {
		parsedMonth, err := time.Parse("2006-01", month)
		if err != nil {
			return helpers.JSONResponse(c, fiber.StatusBadRequest, "Format bulan tidak valid", "Bulan harus dalam format YYYY-MM")
		}
		startDate := parsedMonth
		endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
		query = query.Where("sale_date BETWEEN ? AND ?", startDate, endDate)
	}

	// Optional search by sales.id (tetap case-insensitive seperti semula)
	if search != "" {
		searchLower := strings.ToLower(search)
		query = query.Where("LOWER(id) LIKE ?", "%"+searchLower+"%")
	}

	query = query.Order("sale_date DESC")

	if err := query.Scan(&sales).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data penjualan", err.Error())
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Data penjualan berhasil diambil", sales)
}
