package services

import (
	"errors"
	"time"

	gorm "gorm.io/gorm"
)

var ErrExpenseDataExpiredToEdit = errors.New("data tidak bisa diedit karena sudah tersimpan lebih dari 1 jam")

func ParseExpenseDate(inputDate string, fallback time.Time) (time.Time, error) {
	if inputDate == "" {
		return fallback, nil
	}
	return time.Parse("2006-01-02", inputDate)
}

func NormalizeExpensePayment(current, incoming string) string {
	if incoming == "" {
		return current
	}
	return incoming
}

func EnsureExpenseEditable(db *gorm.DB, expenseID string) error {
	editable, err := IsEditable(db, "expenses", expenseID, 1*time.Hour)
	if err != nil {
		return err
	}
	if !editable {
		return ErrExpenseDataExpiredToEdit
	}
	return nil
}
