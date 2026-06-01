package services

import (
	"errors"
	"time"

	configs "github.com/heru-oktafian/fiber-apotek/configs"
	models "github.com/heru-oktafian/fiber-apotek/models"
	gorm "gorm.io/gorm"
)

// Insert atau update laporan transaksi berdasarkan Sale
func SyncSaleReport(db *gorm.DB, sale models.Sales) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	// Siapkan data report dari Sale
	report := models.TransactionReports{
		ID:              sale.ID,
		TransactionType: models.Sale,
		UserID:          sale.UserID,
		BranchID:        sale.BranchID,
		Total:           sale.TotalSale,
		CreatedAt:       sale.CreatedAt,
		UpdatedAt:       sale.UpdatedAt,
		Payment:         sale.Payment,
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

func RecalculateTotalSale(db *gorm.DB, saleID string) error {
	var total int
	var profitEstimate int

	// Ambil seluruh item dari sale
	var saleItems []models.SaleItems
	if err := db.Where("sale_id = ?", saleID).Find(&saleItems).Error; err != nil {
		return err
	}

	// Ambil data sale termasuk discount
	var sale models.Sales
	if err := db.First(&sale, "id = ?", saleID).Error; err != nil {
		return err
	}

	// Hitung total dan estimasi profit
	for _, item := range saleItems {
		total += item.SubTotal

		var product models.Product
		if err := db.Select("purchase_price").First(&product, "id = ?", item.ProductId).Error; err != nil {
			return err
		}

		profitPerItem := item.Price - product.PurchasePrice
		profitEstimate += profitPerItem * item.Qty
	}

	// Tetapkan diskon (pastikan tidak null)
	discount := sale.Discount

	// Total sale dikurangi diskon
	totalAfterDiscount := total - discount
	if totalAfterDiscount < 0 {
		totalAfterDiscount = 0
	}

	// Estimasi profit juga dikurangi diskon
	finalProfit := profitEstimate - discount
	if finalProfit < 0 {
		finalProfit = 0
	}

	// Update ke tabel sales
	if err := db.Model(&models.Sales{}).Where("id = ?", saleID).Updates(map[string]interface{}{
		"total_sale":      totalAfterDiscount,
		"profit_estimate": finalProfit,
	}).Error; err != nil {
		return err
	}

	// Sync ke laporan transaksi
	if err := SyncSaleReport(db, sale); err != nil {
		return err
	}

	return nil
}
