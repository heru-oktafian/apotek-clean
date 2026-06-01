package services

import (
	"errors"
	"time"

	configs "github.com/heru-oktafian/fiber-apotek/configs"
	models "github.com/heru-oktafian/fiber-apotek/models"
	gorm "gorm.io/gorm"
)

// Insert atau update laporan transaksi berdasarkan DuplicateReceipt
func SyncDuplicateReceiptReport(db *gorm.DB, duplicate_receipt models.DuplicateReceipts) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	// Siapkan data report dari DuplicateReceipt
	report := models.TransactionReports{
		ID:              duplicate_receipt.ID,
		TransactionType: models.Sale,
		UserID:          duplicate_receipt.UserID,
		BranchID:        duplicate_receipt.BranchID,
		Total:           duplicate_receipt.TotalDuplicateReceipt,
		CreatedAt:       duplicate_receipt.CreatedAt,
		UpdatedAt:       duplicate_receipt.UpdatedAt,
		Payment:         duplicate_receipt.Payment,
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

// RecalculateTotalDuplicate menghitung ulang total duplicate receipt
func RecalculateTotalDuplicate(db *gorm.DB, duplicateReceiptID string) error {
	var total int

	// Ambil seluruh item dari duplicate receipt
	var duplicateItems []models.DuplicateReceiptItems
	if err := db.Where("duplicate_receipt_id = ?", duplicateReceiptID).Find(&duplicateItems).Error; err != nil {
		return err
	}

	// Ambil data duplicate receipt termasuk discount
	var duplicateReceipt models.DuplicateReceipts
	if err := db.First(&duplicateReceipt, "id = ?", duplicateReceiptID).Error; err != nil {
		return err
	}

	// Hitung total
	for _, item := range duplicateItems {
		total += item.SubTotal
	}

	// Update ke tabel duplicate_receipts
	if err := db.Model(&models.DuplicateReceipts{}).Where("id = ?", duplicateReceiptID).Updates(map[string]any{
		"total_duplicate_receipt": total,
	}).Error; err != nil {
		return err
	}

	// Sync ke laporan transaksi
	if err := SyncDuplicateReceiptReport(db, duplicateReceipt); err != nil {
		return err
	}

	return nil
}
