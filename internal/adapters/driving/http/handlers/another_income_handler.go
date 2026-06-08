package handlers

import (
	errors "errors"
	math "math"
	strconv "strconv"
	strings "strings"
	time "time"

	configs "apotek-clean/configs"
	helpers "apotek-clean/helpers"
	models "apotek-clean/models"
	services "apotek-clean/services"
	fiber "github.com/gofiber/fiber/v2"
	gorm "gorm.io/gorm"
)

// CreateAnotherIncome Function
func CreateAnotherIncome(c *fiber.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	db := configs.DB

	// Ambil informasi dari token
	branchID, _ := services.GetBranchID(c)
	userID, _ := services.GetUserID(c)
	generatedID := helpers.GenerateID("ANI")

	// Ambil input dari body
	var input models.AnotherIncomeInput
	if err := c.BodyParser(&input); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid input", err)
	}

	parsedDate, err := services.ParseAnotherIncomeDate(input.IncomeDate, nowWIB)
	description := input.Description
	payment := input.Payment
	total := input.TotalIncome
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD", err)
	}

	// Map ke struct model
	another_income := models.AnotherIncomes{
		ID:          generatedID,
		Description: description,
		BranchID:    branchID,
		UserID:      userID,
		IncomeDate:  parsedDate,
		TotalIncome: total,
		Payment:     models.PaymentStatus(payment),
		CreatedAt:   nowWIB,
		UpdatedAt:   nowWIB,
	}

	// Simpan another_income
	if err := db.Create(&another_income).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to create Another Income", err)
	}

	// Buat laporan
	if err := SyncAnotherIncomeReport(db, another_income); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to sync Another Income report", err)
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Another Income created successfully", another_income)
}

// UpdateAnotherIncomeItem Function
func UpdateAnotherIncome(c *fiber.Ctx) error {
	db := configs.DB
	id := c.Params("id")

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	// Cari data another_income
	var another_income models.AnotherIncomes
	if err := db.First(&another_income, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Another Income not found", nil)
	}

	// Gunakan struct khusus input
	var input models.AnotherIncomeInput
	if err := c.BodyParser(&input); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid input", err)
	}

	if err := services.EnsureAnotherIncomeEditable(db, another_income.ID); err != nil {
		if err == services.ErrAnotherIncomeDataExpiredToEdit {
			return helpers.JSONResponse(c, fiber.StatusForbidden, err.Error(), nil)
		}
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to retrieve another income timestamp", err)
	}

	if input.IncomeDate != "" {
		parsedDate, err := services.ParseAnotherIncomeDate(input.IncomeDate, nowWIB)
		if err != nil {
			return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD", err)
		}
		another_income.IncomeDate = parsedDate
	}

	if input.Description != "" {
		another_income.Description = input.Description
	}
	if input.TotalIncome != 0 {
		another_income.TotalIncome = input.TotalIncome
	}
	another_income.Payment = models.PaymentStatus(services.NormalizeAnotherIncomePayment(string(another_income.Payment), input.Payment))
	another_income.UpdatedAt = nowWIB

	// Simpan update
	if err := db.Save(&another_income).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to update Another Income", err)
	}

	// Sync report
	if err := SyncAnotherIncomeReport(db, another_income); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to sync Another Income report", err)
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Another Income updated successfully", another_income)
}

// DeleteAnotherIncomeItem Function
func DeleteAnotherIncome(c *fiber.Ctx) error {
	db := configs.DB
	id := c.Params("id")

	// Ambil another_income
	var another_income models.AnotherIncomes
	if err := db.First(&another_income, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Another Income not found", nil)
	}

	// Hapus laporan
	if err := db.Where("id = ? AND transaction_type = ?", another_income.ID, models.Income).Delete(&models.TransactionReports{}).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to delete transaction report", err)
	}

	// Hapus another_income
	if err := db.Delete(&another_income).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to delete Another Income", err)
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Another Income deleted successfully", another_income)
}

// GetAllAnotherIncome tampilkan semua AnotherIncome
func GetAllAnotherIncomes(c *fiber.Ctx) error {
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

	var AnotherIncome []models.AnotherIncomes
	var total int64

	// Buat builder kueri yang bersih untuk menghitung dan mengambil data
	countQuery := configs.DB.Table("another_incomes ex").
		Where("ex.branch_id = ?", branchID)

	dataQuery := configs.DB.Table("another_incomes ex").
		Select("ex.id, ex.description, ex.income_date, ex.total_income, ex.payment").
		Where("ex.branch_id = ?", branchID)

	// Terapkan filter pencarian
	if search != "" {
		search = strings.ToLower(search)
		countQuery = countQuery.Where("LOWER(ex.description) ILIKE ? ", "%"+search+"%")
		dataQuery = dataQuery.Where("LOWER(ex.description) ILIKE ? ", "%"+search+"%")
	}

	// Terapkan filter bulan
	if month != "" {
		parsedMonth, err := time.Parse("2006-01", month)
		if err != nil {
			return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid month format", err)
		}
		startDate := parsedMonth
		endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
		countQuery = countQuery.Where("ex.income_date BETWEEN ? AND ?", startDate, endDate)
		dataQuery = dataQuery.Where("ex.income_date BETWEEN ? AND ?", startDate, endDate)
	}

	// Pertama, hitung total catatan yang sesuai dengan filter
	if err := countQuery.Count(&total).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to count another income", err)
	}

	// Kemudian, ambil data yang dipaginasi dengan pengurutan
	if err := dataQuery.Order("ex.created_at DESC").Limit(limit).Offset(offset).Find(&AnotherIncome).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to get another income data", err)
	}

	// Buat slice baru untuk menampung data yang sudah diformat
	var formattedAnotherIncomesData []models.AnotherIncomeDetailResponse
	for _, anothIn := range AnotherIncome {
		formattedAnotherIncomesData = append(formattedAnotherIncomesData, models.AnotherIncomeDetailResponse{
			ID:          anothIn.ID,
			Description: anothIn.Description,
			IncomeDate:  helpers.FormatIndonesianDate(anothIn.IncomeDate), // Format tanggal di sini
			TotalIncome: anothIn.TotalIncome,
			Payment:     string(anothIn.Payment),
		})
	}

	// Hitung total halaman berdasarkan hasil filter
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Another Incomes retrieved successfully", search, int(total), page, int(totalPages), int(limit), formattedAnotherIncomesData)
}

// Insert atau update laporan transaksi berdasarkan Another Income / Pendapatan Lain
func SyncAnotherIncomeReport(db *gorm.DB, anotherIncome models.AnotherIncomes) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	// Siapkan data report dari AnotherIncome
	report := models.TransactionReports{
		ID:              anotherIncome.ID,
		TransactionType: models.Income,
		UserID:          anotherIncome.UserID,
		BranchID:        anotherIncome.BranchID,
		Total:           anotherIncome.TotalIncome,
		CreatedAt:       anotherIncome.CreatedAt,
		UpdatedAt:       anotherIncome.UpdatedAt,
		Payment:         anotherIncome.Payment,
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
