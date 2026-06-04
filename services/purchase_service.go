package services

import (
	"errors"
	"time"

	models "apotek-clean/models"
	gorm "gorm.io/gorm"
)

var ErrDataExpiredToEdit = errors.New("data tidak bisa diedit karena sudah tersimpan lebih dari 1 jam")

func EnsurePurchaseEditable(db *gorm.DB, purchaseID string) error {
	editable, err := IsEditable(db, "purchases", purchaseID, 1*time.Hour)
	if err != nil {
		return err
	}
	if !editable {
		return ErrDataExpiredToEdit
	}
	return nil
}

func SumPurchaseItems(items []models.PurchaseItems) int {
	total := 0
	for _, item := range items {
		total += item.SubTotal
	}
	return total
}

func ParsePurchaseDate(inputDate string, fallback time.Time) (time.Time, error) {
	if inputDate == "" {
		return fallback, nil
	}
	return time.Parse("2006-01-02", inputDate)
}

type PreparedPurchaseItem struct {
	ConversionValue int
	ActualQtyToAdd  int
	ItemPrice       int
	ItemSubTotal    int
}

func PreparePurchaseItemValues(qty, price, conversionValue int) PreparedPurchaseItem {
	if conversionValue <= 0 {
		conversionValue = 1
	}
	return PreparedPurchaseItem{
		ConversionValue: conversionValue,
		ActualQtyToAdd:  qty * conversionValue,
		ItemPrice:       price * conversionValue,
		ItemSubTotal:    (price * conversionValue) * qty,
	}
}

type PurchaseItemLookupResult struct {
	Product         models.Product
	Unit            models.Unit
	ConversionValue int
}

func LookupPurchaseItemDependencies(db *gorm.DB, branchID, productID, unitID string) (PurchaseItemLookupResult, error) {
	var result PurchaseItemLookupResult

	if err := db.Where("id = ?", productID).First(&result.Product).Error; err != nil {
		return result, err
	}

	if err := db.Where("id = ?", unitID).First(&result.Unit).Error; err != nil {
		return result, err
	}

	result.ConversionValue = 1
	if unitID != result.Product.UnitId {
		var unitConversion models.UnitConversion
		err := db.Where("product_id = ? AND init_id = ? AND final_id = ? AND branch_id = ?",
			productID,
			unitID,
			result.Product.UnitId,
			branchID,
		).First(&unitConversion).Error
		if err == nil {
			result.ConversionValue = unitConversion.ValueConv
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return result, err
		}
	}

	return result, nil
}
