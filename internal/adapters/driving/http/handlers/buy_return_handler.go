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

func rollbackBuyReturnWithJSON(c *fiber.Ctx, tx *gorm.DB, status int, message string, err interface{}) error {
	tx.Rollback()
	return helpers.JSONResponse(c, status, message, err)
}

func createBuyReturnTransactionReport(tx *gorm.DB, buyReturn models.BuyReturns, userID, branchID string, nowWIB time.Time) error {
	transactionReport := models.TransactionReports{
		ID:              helpers.GenerateID("TRX"),
		TransactionType: models.BuyReturn,
		UserID:          userID,
		BranchID:        branchID,
		Total:           buyReturn.TotalReturn,
		Payment:         buyReturn.Payment,
		CreatedAt:       nowWIB,
		UpdatedAt:       nowWIB,
	}
	return tx.Create(&transactionReport).Error
}

func applyBuyReturnQuotaIfNeeded(tx *gorm.DB, subscriptionType, branchID string) error {
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

// CreateBuyReturnTransaction adalah fungsi untuk membuat transaksi retur pembelian baru
func CreateBuyReturnTransaction(c *fiber.Ctx) error {
	nowWIB := time.Now().In(configs.Location)

	subscriptionType, _ := services.GetClaimsToken(c, "subscription_type")
	branchID, _ := services.GetClaimsToken(c, "branch_id")
	userID, _ := services.GetClaimsToken(c, "user_id")

	db := configs.DB
	var req models.BuyReturnRequest // pastikan struct request ini ada
	err := c.BodyParser(&req)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Body permintaan tidak valid", err.Error())
	}

	if req.BuyReturn.Payment == "" {
		req.BuyReturn.Payment = "paid_by_cash"
	}

	req.BuyReturn.UserID = userID
	req.BuyReturn.BranchID = branchID

	err = helpers.ValidateStruct(req.BuyReturn)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Validasi input retur pembelian gagal", err.Error())
	}

	for _, item := range req.BuyReturnItems {
		err = helpers.ValidateStruct(item)
		if err != nil {
			return helpers.JSONResponse(c, fiber.StatusBadRequest, "Validasi salah satu item retur pembelian gagal", err.Error())
		}
	}

	returnDate, err := services.ParseBuyReturnDate(req.BuyReturn.ReturnDate, nowWIB)
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

	// Validasi apakah purchase_id valid
	var buy models.Purchases
	err = tx.Where("id = ?", req.BuyReturn.PurchaseId).First(&buy).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return rollbackBuyReturnWithJSON(c, tx, fiber.StatusNotFound, fmt.Sprintf("Pembelian dengan ID %s tidak ditemukan", req.BuyReturn.PurchaseId), err.Error())
		}
		return rollbackBuyReturnWithJSON(c, tx, fiber.StatusInternalServerError, "Gagal mengambil data pembelian", err.Error())
	}

	buyReturnID := helpers.GenerateID("BRT")
	buyReturn := models.BuyReturns{
		ID:         buyReturnID,
		PurchaseId: req.BuyReturn.PurchaseId,
		ReturnDate: returnDate,
		BranchID:   branchID,
		Payment:    req.BuyReturn.Payment,
		UserID:     userID,
		CreatedAt:  nowWIB,
		UpdatedAt:  nowWIB,
	}

	var totalReturn int
	var buyReturnItems []models.BuyReturnItems

	for _, item := range req.BuyReturnItems {
		parsedExpiredDate, err := time.Parse("2006-01-02", item.ExpiredDate)
		if err != nil {
			return rollbackBuyReturnWithJSON(c, tx, fiber.StatusBadRequest, fmt.Sprintf("expired_date tidak valid untuk produk %s", item.ProductId), err.Error())
		}

		lookup, err := services.LookupBuyReturnPurchaseItem(tx, req.BuyReturn.PurchaseId, item.ProductId)
		if err != nil {
			return rollbackBuyReturnWithJSON(c, tx, fiber.StatusBadRequest, fmt.Sprintf("Produk %s tidak ditemukan pada pembelian asal", item.ProductId), err.Error())
		}

		totalReturnedQty, err := services.LookupBuyReturnReturnedQty(tx, req.BuyReturn.PurchaseId, item.ProductId)
		if err != nil {
			return rollbackBuyReturnWithJSON(c, tx, fiber.StatusInternalServerError, "Gagal memeriksa retur sebelumnya", err.Error())
		}
		buyItem := lookup.PurchaseItem

		if err := services.ValidateBuyReturnQuantity(buyItem.Qty, item.Qty, totalReturnedQty, item.ProductId); err != nil {
			return rollbackBuyReturnWithJSON(c, tx, fiber.StatusBadRequest, fmt.Sprintf("Total qty retur untuk produk %s melebihi jumlah yang dibeli. Dibeli: %d, Sudah Diretur: %d, Retur Ini: %d",
				item.ProductId, buyItem.Qty, totalReturnedQty, item.Qty), nil)
		}

		// Ambil informasi produk
		var product models.Product
		err = tx.Where("id = ?", item.ProductId).First(&product).Error
		if err != nil {
			return rollbackBuyReturnWithJSON(c, tx, fiber.StatusInternalServerError, fmt.Sprintf("Gagal mengambil info produk untuk %s", item.ProductId), err.Error())
		}

		conversionValue := 1

		// Lakukan konversi unit jika diperlukan
		if buyItem.UnitId != product.UnitId {
			var unitConv models.UnitConversion
			err = tx.Where("product_id = ? AND init_id = ? AND final_id = ? AND branch_id = ?",
				buyItem.ProductId, buyItem.UnitId, product.UnitId, branchID).First(&unitConv).Error

			if err != nil {
				if err != gorm.ErrRecordNotFound {
					return rollbackBuyReturnWithJSON(c, tx, fiber.StatusInternalServerError, "Gagal mengambil konversi satuan", err.Error())
				}
			} else if unitConv.ValueConv > 0 {
				conversionValue = unitConv.ValueConv
			}
		}

		preparedItem := services.PrepareBuyReturnItem(helpers.GenerateID("BRI"), buyReturnID, item, buyItem.Price, conversionValue, parsedExpiredDate)

		// Update stok
		err = tx.Model(&models.Product{}).Where("id = ?", item.ProductId).
			Update("stock", gorm.Expr("stock - ?", preparedItem.ActualQtyToReduce)).Error
		if err != nil {
			return rollbackBuyReturnWithJSON(c, tx, fiber.StatusInternalServerError, fmt.Sprintf("Gagal memperbarui stok untuk produk %s", item.ProductId), err.Error())
		}

		totalReturn += preparedItem.SubTotal
		buyReturnItems = append(buyReturnItems, preparedItem.BuyReturnItem)

	}

	buyReturn.TotalReturn = totalReturn

	err = tx.Create(&buyReturn).Error
	if err != nil {
		return rollbackBuyReturnWithJSON(c, tx, fiber.StatusInternalServerError, "Gagal membuat retur pembelian", err.Error())
	}

	err = tx.CreateInBatches(&buyReturnItems, len(buyReturnItems)).Error
	if err != nil {
		return rollbackBuyReturnWithJSON(c, tx, fiber.StatusInternalServerError, "Gagal membuat item retur pembelian", err.Error())
	}

	err = createBuyReturnTransactionReport(tx, buyReturn, userID, branchID, nowWIB)
	if err != nil {
		return rollbackBuyReturnWithJSON(c, tx, fiber.StatusInternalServerError, "Gagal membuat laporan transaksi retur pembelian", err.Error())
	}

	err = applyBuyReturnQuotaIfNeeded(tx, subscriptionType, branchID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return rollbackBuyReturnWithJSON(c, tx, fiber.StatusInternalServerError, "Gagal mengambil info cabang untuk kuota", err.Error())
		}
		if err.Error() == "quota exhausted" {
			return rollbackBuyReturnWithJSON(c, tx, fiber.StatusBadRequest, "Kuota cabang sudah habis", "Silakan tingkatkan paket berlangganan Anda")
		}
		return rollbackBuyReturnWithJSON(c, tx, fiber.StatusInternalServerError, "Gagal memperbarui kuota cabang", err.Error())
	}

	err = tx.Commit().Error
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal melakukan commit transaksi", err.Error())
	}

	response := services.BuildBuyReturnResponse(buyReturn, buyReturnItems)
	response["return_date"] = helpers.FormatIndonesianDate(buyReturn.ReturnDate)
	return helpers.JSONResponse(c, fiber.StatusOK, "Transaksi retur pembelian berhasil dibuat", response)
}

// GetBuyItemsForReturn digunakan untuk mengambil item pembelian yang bisa diretur
func GetBuyItemsForReturn(c *fiber.Ctx) error {
	purchaseId := c.Query("purchase_id")
	if purchaseId == "" {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "purchase_id wajib diisi", nil)
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
        FROM purchase_items A
        LEFT JOIN products B ON B.id = A.product_id
        LEFT JOIN units C ON C.id = B.unit_id
        LEFT JOIN (
            SELECT 
                sri.product_id, 
                SUM(sri.qty) AS total_returned
            FROM buy_return_items sri
            INNER JOIN buy_returns sr ON sri.buy_return_id = sr.id
            WHERE sr.purchase_id = ?
            GROUP BY sri.product_id
        ) R ON R.product_id = A.product_id
        WHERE A.purchase_id = ? 
        AND COALESCE(R.total_returned, 0) < A.qty
    `, purchaseId, purchaseId).Scan(&results).Error

	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil item pembelian", err.Error())
	}

	if len(results) == 0 {
		return helpers.JSONResponse(c, fiber.StatusOK, "Tidak ada item yang bisa diretur untuk pembelian ini", results)
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Data item retur ditemukan", results)
}

// GetBuyReturnWithItems menampilkan satu retur pembelian beserta semua item-nya
func GetBuyReturnWithItems(c *fiber.Ctx) error {
	db := configs.DB

	buyReturnID := c.Params("id")

	// Gunakan models.AllBuyReturns untuk mengambil data dari DB
	var buyReturn models.AllBuyReturns

	err := db.Table("buy_returns A").
		Select("A.id, A.purchase_id, A.return_date, A.payment, A.total_return").
		Where("A.id = ?", buyReturnID).
		Scan(&buyReturn).Error

	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data retur pembelian", err.Error())
	}

	// Ambil item retur pembelian terkait
	var items []models.AllBuyReturnItems
	err = db.Table("buy_return_items A").
		Select("A.id, A.buy_return_id, A.product_id AS pro_id, B.name AS pro_name, B.unit_id, C.name AS unit_name, A.qty, A.price, A.sub_total, A.expired_date").
		Joins("LEFT JOIN products B on B.id=A.product_id").
		Joins("LEFT JOIN units C on C.id=B.unit_id").
		Where("A.buy_return_id = ?", buyReturnID).
		Order("B.name ASC").
		Scan(&items).Error

	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil item retur pembelian", err.Error())
	}

	// Format tanggal secara manual untuk respons ini
	formattedBuyReturnDate := helpers.FormatIndonesianDate(buyReturn.ReturnDate)

	// Buat objek respons menggunakan struct BuyItemResponse yang baru
	responseDetail := models.BuyReturnItemResponse{
		ID:          buyReturn.ID,
		PurchaseId:  buyReturn.PurchaseId,
		ReturnDate:  formattedBuyReturnDate,
		TotalReturn: buyReturn.TotalReturn,
		Payment:     string(buyReturn.Payment),
		Items:       items,
	}

	// Panggil JSONResponse yang sudah ada, meneruskan BuyItemResponse sebagai 'data'
	return helpers.JSONResponse(c, fiber.StatusOK, "Retur pembelian berhasil diambil", responseDetail)
}

// GetAllBuyReturns menampilkan semua retur pembelian
func GetAllBuyReturns(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)

	var buyReturnsFromDB []models.AllBuyReturns

	query := configs.DB.Table("buy_returns A").
		Select("A.id, A.purchase_id, A.return_date, A.payment, A.total_return").
		Where("A.branch_id = ? ", branchID).
		Order("A.created_at DESC")

	// panggil helper paginate dengan parameter search dan month
	_, search, total, page, totalPages, err := helpers.PaginateWithSearchAndMonth(
		c,
		query,
		&buyReturnsFromDB,
		[]string{"A.purchase_id"},
		"A.return_date",
		1,
		10,
	)

	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil retur pembelian", err.Error())
	}

	// Buat slice baru untuk menampung data yang sudah diformat
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

// CmbPurchase mengambil data pembelian
func CmbPurchase(c *fiber.Ctx) error {
	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	branchID, _ := services.GetBranchID(c)

	// Ambil parameter page dan search dari query URL
	search := strings.TrimSpace(c.Query("search"))

	month := strings.TrimSpace(c.Query("month"))

	// Jika month kosong, isi dengan bulan ini (format YYYY-MM)
	if month == "" {
		month = nowWIB.Format("2006-01")
	}

	// Gunakan struct ringan hanya dengan kolom yang dibutuhkan
	var purchases []struct {
		ID            string    `json:"id"`
		PurchaseDate  time.Time `json:"purchase_date"`
		SupplierName  string    `json:"supplier_name"`
		TotalPurchase int       `json:"total_purchase"`
	}

	query := configs.DB.Table("purchases").
		Select("purchases.id, purchases.purchase_date, suppliers.name AS supplier_name, purchases.total_purchase").
		Joins("LEFT JOIN suppliers ON suppliers.id = purchases.supplier_id").
		Where("purchases.branch_id = ?", branchID)

	// Filter by month (purchase_date)
	if month != "" {
		parsedMonth, err := time.Parse("2006-01", month)
		if err != nil {
			return helpers.JSONResponse(c, fiber.StatusBadRequest, "Format bulan tidak valid", "Bulan harus dalam format YYYY-MM")
		}
		startDate := parsedMonth
		endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
		query = query.Where("purchase_date BETWEEN ? AND ?", startDate, endDate)
	}

	// Optional search by purchases.id (tetap case-insensitive)
	if search != "" {
		searchLower := strings.ToLower(search)
		query = query.Where("LOWER(purchases.id) LIKE ?", "%"+searchLower+"%")
	}

	query = query.Order("purchases.purchase_date DESC")

	if err := query.Scan(&purchases).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data pembelian", err.Error())
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Data pembelian berhasil diambil", purchases)
}
