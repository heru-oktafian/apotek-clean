package services

import (
	"errors"
	"time"

	models "apotek-clean/models"
	gorm "gorm.io/gorm"
)

var ErrDuplicateReceiptDataExpiredToEdit = errors.New("data tidak bisa diedit karena sudah tersimpan lebih dari 1 jam")

type PreparedDuplicateReceiptTotals struct {
	Total  int
	Profit int
}

func EnsureDuplicateReceiptEditable(db *gorm.DB, duplicateReceiptID string) error {
	editable, err := IsEditable(db, "duplicate_receipts", duplicateReceiptID, 1*time.Hour)
	if err != nil {
		return err
	}
	if !editable {
		return ErrDuplicateReceiptDataExpiredToEdit
	}
	return nil
}

func AddDuplicateReceiptContribution(current PreparedDuplicateReceiptTotals, price, purchasePrice, qty, subTotal int) PreparedDuplicateReceiptTotals {
	current.Total += subTotal
	current.Profit += (price - purchasePrice) * qty
	return current
}

type DuplicateReceiptProductLookup struct {
	Product models.Product
}

func LookupDuplicateReceiptProduct(db *gorm.DB, productID string) (DuplicateReceiptProductLookup, error) {
	var result DuplicateReceiptProductLookup
	err := db.Where("id = ?", productID).First(&result.Product).Error
	return result, err
}

func ValidateDuplicateReceiptStock(product models.Product, qty int) error {
	if product.Stock < qty {
		return errors.New("insufficient stock")
	}
	return nil
}

type DuplicateReceiptStockUpdate struct {
	NewStock int
}

func BuildDuplicateReceiptStockUpdate(currentStock, qty int) DuplicateReceiptStockUpdate {
	return DuplicateReceiptStockUpdate{NewStock: currentStock - qty}
}

type PreparedDuplicateReceiptItem struct {
	UpdatedStock DuplicateReceiptStockUpdate
	Totals       PreparedDuplicateReceiptTotals
}

func PrepareDuplicateReceiptItem(item models.DuplicateReceiptItems, lookup DuplicateReceiptProductLookup, runningTotals PreparedDuplicateReceiptTotals) PreparedDuplicateReceiptItem {
	updatedStock := BuildDuplicateReceiptStockUpdate(lookup.Product.Stock, item.Qty)
	updatedTotals := AddDuplicateReceiptContribution(runningTotals, item.Price, lookup.Product.PurchasePrice, item.Qty, item.SubTotal)
	return PreparedDuplicateReceiptItem{
		UpdatedStock: updatedStock,
		Totals:       updatedTotals,
	}
}

func SumDuplicateReceiptItemTotals(items []models.DuplicateReceiptItems) PreparedDuplicateReceiptTotals {
	totals := PreparedDuplicateReceiptTotals{}
	for _, item := range items {
		totals.Total += item.SubTotal
		totals.Profit += item.SubTotal - item.Price*item.Qty
	}
	return totals
}

func ResolveDuplicateReceiptMemberID(db *gorm.DB, inputMemberID, defaultMemberID string) string {
	if inputMemberID == "" {
		return ""
	}
	var member models.Member
	if err := db.Where("id = ?", inputMemberID).First(&member).Error; err != nil {
		return defaultMemberID
	}
	return defaultMemberID
}
