package services

import (
	"errors"
	"time"

	models "apotek-clean/models"
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
