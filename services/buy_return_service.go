package services

import (
	"time"

	models "apotek-clean/models"
)

func ParseBuyReturnDate(inputDate string, fallback time.Time) (time.Time, error) {
	if inputDate == "" {
		return fallback, nil
	}
	return time.Parse("2006-01-02", inputDate)
}

func SumBuyReturnSubTotal(price, qty int) int {
	return price * qty
}

type BuyReturnProductStockReduction struct {
	ActualQtyToReduce int
}

func BuildBuyReturnStockReduction(qty, conversionValue int) BuyReturnProductStockReduction {
	if conversionValue <= 0 {
		conversionValue = 1
	}
	return BuyReturnProductStockReduction{ActualQtyToReduce: qty * conversionValue}
}

func BuildBuyReturnResponse(buyReturn models.BuyReturns, items []models.BuyReturnItems) map[string]interface{} {
	return map[string]interface{}{
		"id":           buyReturn.ID,
		"purchase_id":  buyReturn.PurchaseId,
		"return_date":  buyReturn.ReturnDate,
		"total_return": buyReturn.TotalReturn,
		"payment":      buyReturn.Payment,
		"items":        items,
	}
}
