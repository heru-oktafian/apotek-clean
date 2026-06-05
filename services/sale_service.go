package services

import (
	"errors"
	"time"

	models "apotek-clean/models"
	gorm "gorm.io/gorm"
)

var ErrSaleDataExpiredToEdit = errors.New("data tidak bisa diedit karena sudah tersimpan lebih dari 1 jam")

func EnsureSaleEditable(db *gorm.DB, saleID string) error {
	editable, err := IsEditable(db, "sales", saleID, 1*time.Hour)
	if err != nil {
		return err
	}
	if !editable {
		return ErrSaleDataExpiredToEdit
	}
	return nil
}

type PreparedSaleTotals struct {
	TotalSale      int
	ProfitEstimate int
}

func AddSaleItemContribution(current PreparedSaleTotals, price, purchasePrice, qty, subTotal int) PreparedSaleTotals {
	current.TotalSale += subTotal
	current.ProfitEstimate += (price - purchasePrice) * qty
	return current
}

type SaleItemStockUpdate struct {
	NewStock int
}

func BuildSaleItemStockUpdate(currentStock, qty int) SaleItemStockUpdate {
	return SaleItemStockUpdate{NewStock: currentStock - qty}
}

type SaleProductLookup struct {
	Product models.Product
}

func LookupSaleProduct(db *gorm.DB, productID string) (SaleProductLookup, error) {
	var result SaleProductLookup
	err := db.Where("id = ?", productID).First(&result.Product).Error
	return result, err
}

func ValidateSaleStock(product models.Product, qty int) error {
	if product.Stock < qty {
		return errors.New("insufficient stock")
	}
	return nil
}
