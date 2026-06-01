package services

import (
	errors "errors"
	log "log"
	time "time"

	configs "github.com/heru-oktafian/fiber-apotek/configs"
	models "github.com/heru-oktafian/fiber-apotek/models"
	gorm "gorm.io/gorm"
)

// Insert atau update laporan transaksi berdasarkan Purchase
func SyncPurchaseReport(db *gorm.DB, purchase models.Purchases) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	// Siapkan data report dari purchase
	report := models.TransactionReports{
		ID:              purchase.ID,
		TransactionType: models.Purchase,
		UserID:          purchase.UserID,
		BranchID:        purchase.BranchID,
		Total:           purchase.TotalPurchase,
		CreatedAt:       purchase.CreatedAt,
		UpdatedAt:       purchase.UpdatedAt,
		Payment:         purchase.Payment,
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

// AutoCleanupPurchases will delete any purchases older than 2 hours without purchase items
func AutoCleanupPurchases(db *gorm.DB) error {
	var purchases []models.Purchases

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	// Ambil semua purchase yang tidak punya purchase_items dan lebih dari 2 jam
	err := db.
		Where("created_at < ?", nowWIB.Add(-2*time.Hour)).
		Find(&purchases).Error
	if err != nil {
		return err
	}

	for _, purchase := range purchases {
		var itemCount int64
		db.Model(&models.PurchaseItems{}).
			Where("purchase_id = ?", purchase.ID).
			Count(&itemCount)

		if itemCount == 0 {
			log.Printf("🧹 Auto-cleaning orphan purchase: %s\n", purchase.ID)

			// Hapus transaction_report
			db.Where("id = ?", purchase.ID).Delete(&models.TransactionReports{})

			// Hapus purchase
			db.Where("id = ?", purchase.ID).Delete(&models.Purchases{})
		}
	}

	return nil
}

// RecalculateTotalPurchase menghitung ulang total pembelian berdasarkan item
func RecalculateTotalPurchase(db *gorm.DB, purchaseID string) error {
	var total int64

	// Hitung total sub_total dari purchase_items
	err := db.Model(&models.PurchaseItems{}).
		Where("purchase_id = ?", purchaseID).
		Select("COALESCE(SUM(sub_total), 0)").
		Scan(&total).Error

	if err != nil {
		return err
	}

	// Update ke purchases
	if err := db.Model(&models.Purchases{}).
		Where("id = ?", purchaseID).
		Update("total_purchase", total).Error; err != nil {
		return err
	}

	// Ambil purchase lengkap buat update report
	var purchase models.Purchases
	if err := db.First(&purchase, "id = ?", purchaseID).Error; err != nil {
		return err
	}

	// Update transaction_reports juga
	if err := SyncPurchaseReport(db, purchase); err != nil {
		return err
	}

	return nil
}
