package services

import (
	"errors"
	"time"

	models "apotek-clean/models"
	gorm "gorm.io/gorm"
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

type BuyReturnQtyValidation struct {
	PreviouslyReturned int64
}

func ValidateBuyReturnQuantity(purchasedQty, returningQty int, previousReturned int64, productID string) error {
	if int(previousReturned)+returningQty > purchasedQty {
		return errors.New(productID)
	}
	return nil
}

type BuyReturnPurchaseLookup struct {
	PurchaseItem models.PurchaseItems
	ReturnedQty  int64
}

func LookupBuyReturnPurchaseItem(db *gorm.DB, purchaseID, productID string) (BuyReturnPurchaseLookup, error) {
	var result BuyReturnPurchaseLookup
	err := db.Where("purchase_id = ? AND product_id = ?", purchaseID, productID).First(&result.PurchaseItem).Error
	return result, err
}

func LookupBuyReturnReturnedQty(db *gorm.DB, purchaseID, productID string) (int64, error) {
	var totalReturnedQty int64
	err := db.Table("buy_return_items bri").
		Select("COALESCE(SUM(bri.qty), 0)").
		Joins("JOIN buy_returns br ON br.id = bri.buy_return_id").
		Where("br.purchase_id = ? AND bri.product_id = ?", purchaseID, productID).
		Scan(&totalReturnedQty).Error
	return totalReturnedQty, err
}

type PreparedBuyReturnItem struct {
	BuyReturnItem     models.BuyReturnItems
	ActualQtyToReduce int
	SubTotal          int
}

func PrepareBuyReturnItem(itemID, buyReturnID string, item models.BuyReturnItemInput, price int, conversionValue int, parsedExpiredDate time.Time) PreparedBuyReturnItem {
	reduction := BuildBuyReturnStockReduction(item.Qty, conversionValue)
	subTotal := SumBuyReturnSubTotal(price, item.Qty)
	return PreparedBuyReturnItem{
		BuyReturnItem: models.BuyReturnItems{
			ID:          itemID,
			BuyReturnId: buyReturnID,
			ProductId:   item.ProductId,
			Qty:         item.Qty,
			Price:       price,
			SubTotal:    subTotal,
			ExpiredDate: parsedExpiredDate,
		},
		ActualQtyToReduce: reduction.ActualQtyToReduce,
		SubTotal:          subTotal,
	}
}
