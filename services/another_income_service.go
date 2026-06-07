package services

import "time"

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
