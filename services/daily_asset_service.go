package services

import (
	"math"
	"time"
)

func ParseDailyAssetMonth(input string, fallback time.Time) (string, time.Time, time.Time, error) {
	month := input
	if month == "" {
		month = fallback.Format("2006-01")
	}
	parsedMonth, err := time.Parse("2006-01", month)
	if err != nil {
		return "", time.Time{}, time.Time{}, err
	}
	startDate := parsedMonth
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
	return month, startDate, endDate, nil
}

func CalculateDailyAssetTotalPages(total int64, limit int) int {
	return int(math.Ceil(float64(total) / float64(limit)))
}
