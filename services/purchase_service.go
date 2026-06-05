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

func BuildPurchasedProductUpdates(product models.Product, actualQtyToAdd int, parsedExpiredDate time.Time) map[string]interface{} {
	updates := map[string]interface{}{
		"stock": product.Stock + actualQtyToAdd,
	}

	if parsedExpiredDate.Before(product.ExpiredDate) {
		updates["expired_date"] = parsedExpiredDate
	}

	return updates
}

func BuildPurchaseTransactionResponse(purchase models.Purchases, supplierName string, purchaseDate time.Time, items []models.PurchaseItemResponse) models.PurchaseResponse {
	return models.PurchaseResponse{
		ID:            purchase.ID,
		SupplierID:    purchase.SupplierId,
		SupplierName:  supplierName,
		PurchaseDate:  purchaseDate.Format("02 January 2006"),
		TotalPurchase: purchase.TotalPurchase,
		Payment:       purchase.Payment,
		Items:         items,
	}
}

func LookupPurchaseSupplier(db *gorm.DB, supplierID string) (models.Supplier, error) {
	var supplier models.Supplier
	err := db.Where("id = ?", supplierID).First(&supplier).Error
	return supplier, err
}

type PurchaseItemResponseParams struct {
	Item        models.PurchaseItems
	ProductName string
	UnitName    string
	ExpiredDate time.Time
}

func BuildPurchaseItemResponse(params PurchaseItemResponseParams) models.PurchaseItemResponse {
	return models.PurchaseItemResponse{
		ID:          params.Item.ID,
		ProductID:   params.Item.ProductId,
		ProductName: params.ProductName,
		UnitID:      params.Item.UnitId,
		UnitName:    params.UnitName,
		Price:       params.Item.Price,
		Qty:         params.Item.Qty,
		SubTotal:    params.Item.SubTotal,
		ExpiredDate: params.ExpiredDate.Format("02 January 2006"),
	}
}

type PurchaseItemModelParams struct {
	PurchaseID  string
	ProductID   string
	UnitID      string
	Price       int
	Qty         int
	SubTotal    int
	ExpiredDate time.Time
}

func BuildPurchaseItemModel(id string, params PurchaseItemModelParams) models.PurchaseItems {
	return models.PurchaseItems{
		ID:          id,
		PurchaseId:  params.PurchaseID,
		ProductId:   params.ProductID,
		UnitId:      params.UnitID,
		Price:       params.Price,
		Qty:         params.Qty,
		SubTotal:    params.SubTotal,
		ExpiredDate: params.ExpiredDate,
	}
}
