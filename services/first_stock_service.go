package services

import (
	"errors"
	"time"

	models "apotek-clean/internal/core/entities"
	gorm "gorm.io/gorm"
)

var ErrFirstStockDataExpiredToEdit = errors.New("data tidak bisa diedit karena sudah tersimpan lebih dari 1 jam")

func EnsureFirstStockEditable(db *gorm.DB, firstStockID string) error {
	editable, err := IsEditable(db, "first_stocks", firstStockID, 1*time.Hour)
	if err != nil {
		return err
	}
	if !editable {
		return ErrFirstStockDataExpiredToEdit
	}
	return nil
}

func ParseFirstStockDate(inputDate string, fallback time.Time) (time.Time, error) {
	if inputDate == "" {
		return fallback, nil
	}
	return time.Parse("2006-01-02", inputDate)
}

func SumFirstStockItems(items []models.FirstStockItems) int {
	total := 0
	for _, item := range items {
		total += item.SubTotal
	}
	return total
}

type FirstStockLookup struct {
	Product         models.Product
	Unit            models.Unit
	ConversionValue int
}

func LookupFirstStockDependencies(db *gorm.DB, branchID, productID, unitID string) (FirstStockLookup, error) {
	var result FirstStockLookup
	if err := db.Where("id = ? AND branch_id = ?", productID, branchID).First(&result.Product).Error; err != nil {
		return result, err
	}
	if err := db.Where("id = ?", unitID).First(&result.Unit).Error; err != nil {
		return result, err
	}
	result.ConversionValue = 1
	if unitID != result.Product.UnitId {
		var unitConversion models.UnitConversion
		err := db.Where("product_id = ? AND init_id = ? AND final_id = ? AND branch_id = ?", productID, unitID, result.Product.UnitId, branchID).First(&unitConversion).Error
		if err == nil {
			result.ConversionValue = unitConversion.ValueConv
		} else if err != gorm.ErrRecordNotFound {
			return result, err
		}
	}
	return result, nil
}

type PreparedFirstStockItem struct {
	Item          models.FirstStockItems
	Response      models.FirstStockItemResponse
	ProductUpdate map[string]interface{}
	SubTotal      int
}

func PrepareFirstStockItem(itemID, firstStockID string, input models.FirstStockItemInput, lookup FirstStockLookup, parsedExpiredDate time.Time) PreparedFirstStockItem {
	actualQtyToAdd := input.Qty * lookup.ConversionValue
	itemPrice := lookup.Product.PurchasePrice
	itemSubTotal := itemPrice * actualQtyToAdd
	item := models.FirstStockItems{
		ID:           itemID,
		FirstStockId: firstStockID,
		ProductId:    input.ProductId,
		Price:        itemPrice,
		Qty:          input.Qty,
		SubTotal:     itemSubTotal,
		ExpiredDate:  parsedExpiredDate,
	}
	response := models.FirstStockItemResponse{
		ID:          item.ID,
		ProductID:   item.ProductId,
		ProductName: lookup.Product.Name,
		UnitID:      input.UnitId,
		UnitName:    lookup.Unit.Name,
		Price:       item.Price,
		Qty:         item.Qty,
		SubTotal:    item.SubTotal,
		ExpiredDate: parsedExpiredDate.Format("02 January 2006"),
	}
	updates := map[string]interface{}{"stock": lookup.Product.Stock + actualQtyToAdd}
	if parsedExpiredDate.Before(lookup.Product.ExpiredDate) {
		updates["expired_date"] = parsedExpiredDate
	}
	return PreparedFirstStockItem{Item: item, Response: response, ProductUpdate: updates, SubTotal: itemSubTotal}
}
