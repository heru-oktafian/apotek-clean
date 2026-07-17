package handlers

import (
	strings "strings"
	time "time"

	configs "apotek-clean/configs"
	helpers "apotek-clean/helpers"
	models "apotek-clean/internal/core/entities"
	services "apotek-clean/services"
	reports "apotek-clean/services/reports"
	fiber "github.com/gofiber/fiber/v2"
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
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Input pengeluaran tidak valid", err)
	}

	parsedDate, err := services.ParseExpenseDate(input.ExpenseDate, nowWIB)
	description := input.Description
	payment := input.Payment
	total := input.TotalExpense
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Format tanggal tidak valid. Gunakan YYYY-MM-DD", err)
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
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal membuat pengeluaran", err)
	}

	// Buat laporan
	if err := reports.SyncExpenseReport(db, expense); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal membuat laporan pengeluaran", err)
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Pengeluaran berhasil dibuat", expense)
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
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Pengeluaran tidak ditemukan", err)
	}

	// Gunakan struct khusus input
	var input models.ExpenseInput
	if err := c.BodyParser(&input); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Input pengeluaran tidak valid", err)
	}

	if err := services.EnsureExpenseEditable(db, expense.ID); err != nil {
		if err == services.ErrExpenseDataExpiredToEdit {
			return helpers.JSONResponse(c, fiber.StatusForbidden, err.Error(), nil)
		}
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil timestamp pengeluaran", err)
	}

	if input.ExpenseDate != "" {
		parsedDate, err := services.ParseExpenseDate(input.ExpenseDate, nowWIB)
		if err != nil {
			return helpers.JSONResponse(c, fiber.StatusBadRequest, "Format tanggal tidak valid. Gunakan YYYY-MM-DD", err)
		}
		expense.ExpenseDate = parsedDate
	}

	if input.Description != "" {
		expense.Description = input.Description
	}
	if input.TotalExpense != 0 {
		expense.TotalExpense = input.TotalExpense
	}
	expense.Payment = models.PaymentStatus(services.NormalizeExpensePayment(string(expense.Payment), input.Payment))
	expense.UpdatedAt = nowWIB

	// Simpan update
	if err := db.Save(&expense).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui pengeluaran", err)
	}

	// Sync report
	if err := reports.SyncExpenseReport(db, expense); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal menyinkronkan laporan pengeluaran", err)
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Pengeluaran berhasil diperbarui", expense)
}

// DeleteExpenseItem Function
func DeleteExpense(c *fiber.Ctx) error {
	db := configs.DB
	id := c.Params("id")

	// Ambil expense
	var expense models.Expenses
	if err := db.First(&expense, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Pengeluaran tidak ditemukan", err)
	}

	// Hapus laporan
	if err := db.Where("id = ? AND transaction_type = ?", expense.ID, models.Expense).Delete(&models.TransactionReports{}).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal menghapus laporan transaksi", err)
	}

	// Hapus expense
	if err := db.Delete(&expense).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal menghapus pengeluaran", err)
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Pengeluaran berhasil dihapus", expense)
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
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data pengeluaran", err)
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

	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Data pengeluaran berhasil diambil", search, total, page, totalPages, 10, formattedExpenseData)
}
