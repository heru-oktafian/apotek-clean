package helpers

import (
	errors "errors"
	log "log"
	http "net/http"
	time "time"

	fiber "github.com/gofiber/fiber/v2"
	configs "github.com/heru-oktafian/fiber-apotek/configs"
	models "github.com/heru-oktafian/fiber-apotek/models"
	gorm "gorm.io/gorm"
)

// AutoCleanupOpnames will delete any opnames older than 2 hours without opname items
func AutoCleanupOpnames(db *gorm.DB) error {
	var opnames []models.Opnames

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	// Ambil semua opnames yang tidak punya opname_items
	err := db.
		Where("created_at < ?", nowWIB.Add(-2*time.Hour)).
		Find(&opnames).Error
	if err != nil {
		return err
	}

	for _, opname := range opnames {
		var itemCount int64
		db.Model(&models.OpnameItems{}).
			Where("opname_id = ?", opname.ID).
			Count(&itemCount)

		if itemCount == 0 {
			log.Printf("🧹 Auto-cleaning orphan opname: %s\n", opname.ID)

			// Hapus transaction_report
			db.Where("id = ?", opname.ID).Delete(&models.TransactionReports{})

			// Hapus opname
			db.Where("id = ?", opname.ID).Delete(&models.Opnames{})
		}
	}

	return nil
}

// Insert atau update laporan transaksi berdasarkan FirstStocks / Pengeluaran
func SyncOpnameReport(db *gorm.DB, opname models.Opnames) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	// Siapkan data report dari FirstStock
	report := models.TransactionReports{
		ID:              opname.ID,
		TransactionType: models.Ipname,
		UserID:          opname.UserID,
		BranchID:        opname.BranchID,
		Total:           opname.TotalOpname,
		CreatedAt:       nowWIB,
		UpdatedAt:       nowWIB,
		Payment:         opname.Payment,
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

// JSONFirstStockWithItemsResponse sends a standard JSON response format / structure
func JSONOpnameWithItemsResponse(c *fiber.Ctx, status int, message string, opname_id string, description string, opname_date string, total_opname int, payment string, items interface{}) error {
	resp := models.ResponseOpnameWithItemsResponse{
		Status:      http.StatusText(status),
		Message:     message,
		OpnameId:    opname_id,
		Description: description,
		OpnameDate:  opname_date,
		TotalOpname: total_opname,
		Payment:     payment,
		Items:       items,
	}
	return JSONResponse(c, status, message, resp)
}

// Opname stock product
func OpnameProductStock(db *gorm.DB, productID string, qty int) error {
	var product models.Product
	if err := db.First(&product, "id = ?", productID).Error; err != nil {
		return err
	}
	product.Stock = qty
	return db.Save(&product).Error
}

// RecalculateTotalOpname menghitung ulang total opname
func RecalculateTotalOpname(db *gorm.DB, opnameID string) error {
	var totalAdjustment int64 // Ubah nama variabel untuk merefleksikan 'adjustment'

	// Hitung total opname berdasarkan penjumlahan statis (sub_total - sub_total_exist)
	err := db.Table("opname_items").
		Where("opname_id = ?", opnameID).
		Select("COALESCE(SUM(sub_total - sub_total_exist), 0)").
		Scan(&totalAdjustment).Error                             // Scan ke variabel totalAdjustment

	if err != nil {
		return err
	}

	// Update ke opnames
	// Gunakan totalAdjustment untuk mengupdate kolom total_opname
	if err := db.Model(&models.Opnames{}).
		Where("id = ?", opnameID).
		Update("total_opname", totalAdjustment).Error; err != nil {
		return err
	}

	// Ambil opname lengkap buat update report
	var opname models.Opnames
	if err := db.First(&opname, "id = ?", opnameID).Error; err != nil {
		return err
	}

	// Update transaction_reports juga
	if err := SyncOpnameReport(db, opname); err != nil {
		return err
	}

	return nil
}

// OpnameProductStockAsync set stok produk secara asynchronous
func OpnameProductStockAsync(db *gorm.DB, productID string, qty int) {
	go func() {
		if err := OpnameProductStock(db, productID, qty); err != nil {
			// Log error asynchronously
			log.Printf("Failed to opname product stock asynchronously: %v", err)
		}
	}()
}
