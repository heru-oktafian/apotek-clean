package services

import "time"

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
