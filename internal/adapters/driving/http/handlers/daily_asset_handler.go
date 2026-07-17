package handlers

import (
	"strings"
	"time"

	configs "apotek-clean/configs"
	helpers "apotek-clean/helpers"
	models "apotek-clean/models"
	services "apotek-clean/services"
	fiber "github.com/gofiber/fiber/v2"
)

type DailyAssetHandler struct{}

func NewDailyAssetHandler() *DailyAssetHandler {
	return &DailyAssetHandler{}
}

// GetAllAssets returns all daily assets for the current branch filtered by month (no pagination)
func (h *DailyAssetHandler) GetAllAssets(c *fiber.Ctx) error {
	nowWIB := time.Now().In(configs.Location)
	branchID, _ := services.GetBranchID(c)

	_, startDate, endDate, err := services.ParseDailyAssetMonth(strings.TrimSpace(c.Query("month")), nowWIB)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid month format. Month should be in format YYYY-MM", err)
	}

	// Query all daily assets without pagination
	var dailyAssetFromDB []models.AllDailyAsset
	query := configs.DB.Table("daily_assets ast").
		Select("ast.id, ast.asset_date, ast.asset_value, ast.asset_average, ast.branch_id, bc.branch_name").
		Joins("LEFT JOIN branches bc on bc.id = ast.branch_id").
		Where("ast.branch_id = ?", branchID).
		Where("ast.asset_date BETWEEN ? AND ?", startDate, endDate).
		Order("ast.asset_date DESC")

	if err := query.Scan(&dailyAssetFromDB).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get assets failed", err)
	}

	// Ambil asset_average terbaru pada bulan yang dipilih untuk branch ini
	var latestAvg struct {
		AssetAverage int `gorm:"column:asset_average"`
	}
	if err := configs.DB.Table("daily_assets").
		Select("asset_average").
		Where("branch_id = ? AND asset_date BETWEEN ? AND ?", branchID, startDate, endDate).
		Order("asset_date DESC").
		Limit(1).
		Scan(&latestAvg).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get assets failed", err)
	}

	// Format data dengan tanggal dalam format Indonesia
	var formattedDailyAsset []models.DetailDailyAsset
	for _, daily := range dailyAssetFromDB {
		formattedDailyAsset = append(formattedDailyAsset, models.DetailDailyAsset{
			ID:           daily.ID,
			AssetDate:    helpers.FormatIndonesianDate(daily.AssetDate),
			AssetValue:   daily.AssetValue,
			BranchId:     daily.BranchId,
			AssetAverage: daily.AssetAverage,
			BranchName:   daily.BranchName,
		})
	}

	// Response tanpa pagination
	resp := map[string]interface{}{
		"monthly_asset_average": latestAvg.AssetAverage,
		"data":                 formattedDailyAsset,
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Daily assets retrieved successfully", resp)
}
