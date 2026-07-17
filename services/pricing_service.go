package services

import (
	models "apotek-clean/internal/core/entities"
	gorm "gorm.io/gorm"
)

// UpdateProductPriceIfHigher memperbarui harga produk jika harga baru lebih tinggi
func UpdateProductPriceIfHigher(db *gorm.DB, productId string, newPrice int) error {
	var product models.Product

	if err := db.First(&product, "id = ?", productId).Error; err != nil {
		return err
	}

	if newPrice > product.PurchasePrice {
		product.PurchasePrice = newPrice
		return db.Save(&product).Error
	}

	// Tidak update jika harga baru lebih rendah
	return nil
}
