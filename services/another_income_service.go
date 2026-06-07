package services

import (
	"errors"
	"time"

	gorm "gorm.io/gorm"
)

var ErrAnotherIncomeDataExpiredToEdit = errors.New("data tidak bisa diedit karena sudah tersimpan lebih dari 1 jam")

func ParseAnotherIncomeDate(inputDate string, fallback time.Time) (time.Time, error) {
	if inputDate == "" {
		return fallback, nil
	}
	return time.Parse("2006-01-02", inputDate)
}

func NormalizeAnotherIncomePayment(current, incoming string) string {
	if incoming == "" {
		return current
	}
	return incoming
}

func EnsureAnotherIncomeEditable(db *gorm.DB, incomeID string) error {
	editable, err := IsEditable(db, "another_incomes", incomeID, 1*time.Hour)
	if err != nil {
		return err
	}
	if !editable {
		return ErrAnotherIncomeDataExpiredToEdit
	}
	return nil
}
