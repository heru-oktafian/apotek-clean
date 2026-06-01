package services

import (
	errors "errors"
	fmt "fmt"
	time "time"

	models "github.com/heru-oktafian/fiber-apotek/models"
	gorm "gorm.io/gorm"
)

// Tambah stock product
func AddProductStock(db *gorm.DB, productID string, qty int) error {
	var product models.Product
	if err := db.First(&product, "id = ?", productID).Error; err != nil {
		return err
	}
	product.Stock += qty
	return db.Save(&product).Error
}

// Kurangi stock product
func ReduceProductStock(db *gorm.DB, productID string, qty int) error {
	var product models.Product
	if err := db.First(&product, "id = ?", productID).Error; err != nil {
		return err
	}

	if product.Stock < qty {
		return errors.New("insufficient stock")
	}
	product.Stock -= qty

	if err := db.Save(&product).Error; err != nil {
		return err
	}

	return nil
}

// SubtractProductStock menambah stok produk
func SubtractProductStock(db *gorm.DB, productID string, qty int) error {
	var product models.Product
	if err := db.First(&product, "id = ?", productID).Error; err != nil {
		return err
	}

	product.Stock += qty

	if err := db.Save(&product).Error; err != nil {
		return err
	}

	return nil
}

// ZeroProductStock kosongkan stok produk
func ZeroProductStock(db *gorm.DB, productID string, qty int) error {
	var product models.Product
	if err := db.First(&product, "id = ?", productID).Error; err != nil {
		return err
	}

	product.Stock = 0

	if err := db.Save(&product).Error; err != nil {
		return err
	}

	return nil
}

// AddProductStockAsync menambah stok produk secara asynchronous
func AddProductStockAsync(db *gorm.DB, productID string, qty int) {
	go func() {
		if err := AddProductStock(db, productID, qty); err != nil {
			// Log error asynchronously
			fmt.Printf("Failed to add product stock asynchronously: %v\n", err)
		}
	}()
}

// ReduceProductStockAsync mengurangi stok produk secara asynchronous
func ReduceProductStockAsync(db *gorm.DB, productID string, qty int) {
	go func() {
		if err := ReduceProductStock(db, productID, qty); err != nil {
			// Log error asynchronously
			fmt.Printf("Failed to reduce product stock asynchronously: %v\n", err)
		}
	}()
}

// ZeroProductStockAsync mengosongkan stok produk secara asynchronous
func ZeroProductStockAsync(db *gorm.DB, productID string, qty int) {
	go func() {
		if err := ZeroProductStock(db, productID, qty); err != nil {
			// Log error asynchronously
			fmt.Printf("Failed to zero product stock asynchronously: %v\n", err)
		}
	}()
}

// IsEditable memeriksa apakah sebuah record masih bisa diedit
func IsEditable(db *gorm.DB, tableName string, recordID string, times time.Duration) (bool, error) {
	var createdAt time.Time
	err := db.Table(tableName).
		Select("created_at").
		Where("id = ?", recordID).
		Scan(&createdAt).Error
	if err != nil {
		return false, err
	}

	if time.Since(createdAt) >= times {
		return false, nil
	}

	return true, nil
}
