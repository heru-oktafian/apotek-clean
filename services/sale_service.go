package services

import (
	"errors"
	"time"

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
