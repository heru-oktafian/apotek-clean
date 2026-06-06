package services

import (
	"errors"
	"time"

	models "apotek-clean/models"
	gorm "gorm.io/gorm"
)

func ParseSaleReturnDate(inputDate string, fallback time.Time) (time.Time, error) {
	if inputDate == "" {
		return fallback, nil
	}
	return time.Parse("2006-01-02", inputDate)
}

func SumSaleReturnSubTotal(price, qty int) int {
	return price * qty
}

func ValidateSaleReturnQuantity(soldQty, returningQty int, previousReturned int64, productID string) error {
	if int(previousReturned)+returningQty > soldQty {
		return errors.New(productID)
	}
	return nil
}

type SaleReturnSaleLookup struct {
	SaleItem    models.SaleItems
	ReturnedQty int64
}

func LookupSaleReturnSaleItem(db *gorm.DB, saleID, productID string) (SaleReturnSaleLookup, error) {
	var result SaleReturnSaleLookup
	err := db.Where("sale_id = ? AND product_id = ?", saleID, productID).First(&result.SaleItem).Error
	return result, err
}

func LookupSaleReturnReturnedQty(db *gorm.DB, saleID, productID string) (int64, error) {
	var totalReturnedQty int64
	err := db.Model(&models.SaleReturnItems{}).
		Select("COALESCE(SUM(qty), 0)").
		Where("product_id = ? AND sale_return_id IN (SELECT id FROM sale_returns WHERE sale_id = ?)", productID, saleID).
		Scan(&totalReturnedQty).Error
	return totalReturnedQty, err
}

func BuildSaleReturnResponse(saleReturn models.SaleReturns, items []models.SaleReturnItems) map[string]interface{} {
	return map[string]interface{}{
		"id":           saleReturn.ID,
		"sale_id":      saleReturn.SaleId,
		"return_date":  saleReturn.ReturnDate,
		"total_return": saleReturn.TotalReturn,
		"payment":      saleReturn.Payment,
		"items":        items,
	}
}
