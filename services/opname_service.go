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

func FormatMobileOpnameRows(rows []models.OpnameQueryResult) []models.AllOpnameMobiles {
	formatted := make([]models.AllOpnameMobiles, 0, len(rows))
	for _, op := range rows {
		formatted = append(formatted, models.AllOpnameMobiles{
			ID:          op.ID,
			Description: op.Description,
			OpnameDate:  FormatIndonesianDate(op.OpnameDate),
			TotalOpname: op.TotalOpname,
		})
	}
	return formatted
}
