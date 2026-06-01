package services

import (
	errors "errors"
	time "time"

	configs "github.com/heru-oktafian/fiber-apotek/configs"
	models "github.com/heru-oktafian/fiber-apotek/models"
	gorm "gorm.io/gorm"
)

// Insert atau update laporan transaksi berdasarkan Expenses / Pengeluaran
func SyncExpenseReport(db *gorm.DB, expense models.Expenses) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	// Siapkan data report dari Expense
	report := models.TransactionReports{
		ID:              expense.ID,
		TransactionType: models.Expense,
		UserID:          expense.UserID,
		BranchID:        expense.BranchID,
		Total:           expense.TotalExpense,
		CreatedAt:       expense.CreatedAt,
		UpdatedAt:       expense.UpdatedAt,
		Payment:         expense.Payment,
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
