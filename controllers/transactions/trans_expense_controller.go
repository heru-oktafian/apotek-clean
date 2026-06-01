package controllers

import (
	strings "strings"
	time "time"

	fiber "github.com/gofiber/fiber/v2"
	configs "github.com/heru-oktafian/fiber-apotek/configs"
	helpers "github.com/heru-oktafian/fiber-apotek/helpers"
	models "github.com/heru-oktafian/fiber-apotek/models"
	services "github.com/heru-oktafian/fiber-apotek/services"
	reports "github.com/heru-oktafian/fiber-apotek/services/reports"
)

// CreateExpense Function
func CreateExpense(c *fiber.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	db := configs.DB

	// Ambil informasi dari token
	branchID, _ := services.GetBranchID(c)
	userID, _ := services.GetUserID(c)
	generatedID := helpers.GenerateID("EXP")

	// Ambil input dari body
	var input models.ExpenseInput
	if err := c.BodyParser(&input); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid input", err)
	}

	// Parse tanggal
	layout := "2006-01-02" // format harus YYYY-MM-DD
	parsedDate, err := time.Parse(layout, input.ExpenseDate)
	description := input.Description
	payment := input.Payment
	total := input.TotalExpense
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD", err)
	}

	// Map ke struct model
	expense := models.Expenses{
		ID:           generatedID,
		Description:  description,
		BranchID:     branchID,
		UserID:       userID,
		ExpenseDate:  parsedDate,
		TotalExpense: total,
		Payment:      models.PaymentStatus(payment),
		CreatedAt:    nowWIB,
		UpdatedAt:    nowWIB,
	}

	// Simpan expense
	if err := db.Create(&expense).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to create Expense", err)
	}

	// Buat laporan
	if err := reports.SyncExpenseReport(db, expense); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to create Expense Report", err)
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Expense created successfully", expense)
}

// UpdateExpenseItem Function
func UpdateExpense(c *fiber.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	db := configs.DB
	id := c.Params("id")

	// Cari data expense
	var expense models.Expenses
	if err := db.First(&expense, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Expense not found", err)
	}

	// Gunakan struct khusus input
	var input models.ExpenseInput
	if err := c.BodyParser(&input); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid input", err)
	}

	// Parse tanggal dari string ke time.Time
	layout := "2006-01-02"
	parsedDate, err := time.Parse(layout, input.ExpenseDate)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD", err)
	}

	// Update field dasar
	expense.ExpenseDate = parsedDate
	expense.Description = input.Description
	expense.TotalExpense = input.TotalExpense
	expense.Payment = models.PaymentStatus(input.Payment)
	expense.UpdatedAt = nowWIB

	// Simpan update
	if err := db.Save(&expense).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to update Expense", err)
	}

	// Sync report
	if err := reports.SyncExpenseReport(db, expense); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to sync Expense Report", err)
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Expense updated successfully", expense)
}

// DeleteExpenseItem Function
func DeleteExpense(c *fiber.Ctx) error {
	db := configs.DB
	id := c.Params("id")

	// Ambil expense
	var expense models.Expenses
	if err := db.First(&expense, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Expense not found", err)
	}

	// Hapus laporan
	if err := db.Where("id = ? AND transaction_type = ?", expense.ID, models.Expense).Delete(&models.TransactionReports{}).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to delete Transaction Report", err)
	}

	// Hapus expense
	if err := db.Delete(&expense).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to delete Expense", err)
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Expense deleted successfully", expense)
}

// GetAllExpenses tampilkan semua Expense
func GetAllExpenses(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)

	var expenses []models.Expenses

	// Query dasar
	query := configs.DB.Table("expenses ex").
		Select("ex.id, ex.description, ex.expense_date, ex.total_expense, ex.payment").
		Where("ex.branch_id = ?", branchID).
		Order("ex.created_at DESC")

	// Panggil helper PaginateWithSearchAndMonth
	_, search, total, page, totalPages, err := helpers.PaginateWithSearchAndMonth(
		c,
		query,
		&expenses,
		[]string{"ex.description"}, // Kolom pencarian
		"ex.expense_date",          // Kolom tanggal untuk filter bulan
		1,                          // Default page
		10,                         // Default limit
	)

	if err != nil {
		if strings.Contains(err.Error(), "format bulan tidak valid") {
			return helpers.JSONResponse(c, fiber.StatusBadRequest, err.Error(), nil)
		}
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to get expenses data", err)
	}

	// Format data pengeluaran yang diambil
	var formattedExpenseData []models.ExpenseDetailResponse
	for _, expense := range expenses {
		formattedExpenseData = append(formattedExpenseData, models.ExpenseDetailResponse{
			ID:           expense.ID,
			Description:  expense.Description,
			ExpenseDate:  helpers.FormatIndonesianDate(expense.ExpenseDate),
			TotalExpense: expense.TotalExpense,
			Payment:      string(expense.Payment),
		})
	}

	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Expenses retrieved successfully", search, total, page, totalPages, 10, formattedExpenseData)
}
