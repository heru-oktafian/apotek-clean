package services

import (
	"math"
	"time"

	helpers "apotek-clean/helpers"
	models "apotek-clean/models"
)

func ParseDefectaDate(inputDate string, fallback time.Time) (time.Time, error) {
	if inputDate == "" {
		return fallback, nil
	}
	return time.Parse("2006-01-02", inputDate)
}

func SumDefectaSubTotal(price, qty int) int {
	return price * qty
}

func BuildDefectaItemsResponse(items []models.AllDefectaItems) []models.AllDefectaItems {
	formatted := make([]models.AllDefectaItems, 0, len(items))
	for _, item := range items {
		formatted = append(formatted, models.AllDefectaItems{
			ID:          item.ID,
			ProductId:   item.ProductId,
			ProductName: item.ProductName,
			UnitId:      item.UnitId,
			UnitName:    item.UnitName,
			Price:       item.Price,
			Qty:         item.Qty,
			SubTotal:    item.SubTotal,
		})
	}
	return formatted
}

func CalculateDefectaTotalPages(limit int, total int64) int {
	return int(math.Ceil(float64(total) / float64(limit)))
}

func BuildDefectaListResponse(defectas []models.Defectas) []models.DefectaDetailResponse {
	formatted := make([]models.DefectaDetailResponse, 0, len(defectas))
	for _, defecta := range defectas {
		formatted = append(formatted, models.DefectaDetailResponse{
			ID:            defecta.ID,
			DefectaDate:   helpers.FormatIndonesianDate(defecta.DefectaDate),
			TotalEstimate: defecta.TotalEstimate,
			DefectaStatus: string(defecta.DefectaStatus),
		})
	}
	return formatted
}
