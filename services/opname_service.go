package services

import (
	"time"

	models "apotek-clean/models"
)

func ParseOpnameDate(inputDate string, fallback time.Time) (time.Time, error) {
	if inputDate == "" {
		return fallback, nil
	}
	return time.Parse("2006-01-02", inputDate)
}

func ParseOpnameItemDate(inputDate string) (time.Time, error) {
	return time.Parse("2006-01-02", inputDate)
}

type OpnameItemProductSnapshot struct {
	OldStock          int
	OldPurchasePrice  int
	ParsedExpiredDate time.Time
}

func BuildOpnameItemSnapshot(product models.Product, parsedExpiredDate time.Time) OpnameItemProductSnapshot {
	return OpnameItemProductSnapshot{
		OldStock:          product.Stock,
		OldPurchasePrice:  product.PurchasePrice,
		ParsedExpiredDate: parsedExpiredDate,
	}
}

type PreparedOpnameItem struct {
	Item           models.OpnameItems
	ProductUpdates map[string]interface{}
}

func PrepareOpnameItem(itemID string, input models.CreateOpnameItemInput, snapshot OpnameItemProductSnapshot) PreparedOpnameItem {
	item := models.OpnameItems{
		ID:            itemID,
		OpnameId:      input.OpnameId,
		ProductId:     input.ProductId,
		Qty:           input.Qty,
		ExpiredDate:   snapshot.ParsedExpiredDate,
		Price:         input.Price,
		QtyExist:      snapshot.OldStock,
		SubTotalExist: snapshot.OldStock * snapshot.OldPurchasePrice,
		SubTotal:      input.Qty * input.Price,
	}
	updates := map[string]interface{}{
		"expired_date":   snapshot.ParsedExpiredDate,
		"stock":          input.Qty,
		"purchase_price": input.Price,
	}
	return PreparedOpnameItem{Item: item, ProductUpdates: updates}
}

type PreparedOpnameItemUpdate struct {
	ParsedDate time.Time
	SubTotal   int
}

func PrepareOpnameItemUpdate(expiredDate string, price, qty int) (PreparedOpnameItemUpdate, error) {
	parsedDate, err := ParseOpnameItemDate(expiredDate)
	if err != nil {
		return PreparedOpnameItemUpdate{}, err
	}
	return PreparedOpnameItemUpdate{ParsedDate: parsedDate, SubTotal: price * qty}, nil
}
