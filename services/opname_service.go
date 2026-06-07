package services

import "time"

func ParseOpnameDate(inputDate string, fallback time.Time) (time.Time, error) {
	if inputDate == "" {
		return fallback, nil
	}
	return time.Parse("2006-01-02", inputDate)
}

func ParseOpnameItemDate(inputDate string) (time.Time, error) {
	return time.Parse("2006-01-02", inputDate)
}
