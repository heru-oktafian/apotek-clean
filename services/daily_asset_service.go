package services

import (
	"math"
	"time"

	helpers "apotek-clean/helpers"
	models "apotek-clean/models"
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

func FormatDailyAssetRows(rows []models.AllDailyAsset) []models.DetailDailyAsset {
	formatted := make([]models.DetailDailyAsset, 0, len(rows))
	for _, daily := range rows {
		formatted = append(formatted, models.DetailDailyAsset{
			ID:           daily.ID,
			AssetDate:    helpers.FormatIndonesianDate(daily.AssetDate),
			AssetValue:   daily.AssetValue,
			BranchId:     daily.BranchId,
			AssetAverage: daily.AssetAverage,
			BranchName:   daily.BranchName,
		})
	}
	return formatted
}
