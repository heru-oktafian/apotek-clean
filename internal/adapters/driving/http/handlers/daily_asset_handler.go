package handlers

import (
	strconv "strconv"
	strings "strings"
	time "time"

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

func (h *DailyAssetHandler) GetAllAssets(c *fiber.Ctx) error {
	nowWIB := time.Now().In(configs.Location)
	branchID, _ := services.GetBranchID(c)
	pageParam := c.Query("page")
	page := 1
	if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
		page = p
	}
	limit := 10
	offset := (page - 1) * limit
	month, startDate, endDate, err := services.ParseDailyAssetMonth(strings.TrimSpace(c.Query("month")), nowWIB)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid month format. Month should be in format YYYY-MM", err)
	}
	var dailyAssetFromDB []models.AllDailyAsset
	var total int64
	query := configs.DB.Table("daily_assets ast").
		Select("ast.id, ast.asset_date, ast.asset_value, ast.asset_average, ast.branch_id, bc.branch_name").
		Joins("LEFT JOIN branches bc on bc.id = ast.branch_id").
		Where("ast.branch_id = ? ", branchID).
		Order("ast.asset_date DESC")
	if month != "" {
		query = query.Where("ast.asset_date BETWEEN ? AND ?", startDate, endDate)
	}
	if err := query.Count(&total).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get assets failed", err)
	}
	if err := query.Offset(offset).Limit(limit).Scan(&dailyAssetFromDB).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get assets failed", err)
	}
	var latestAvg struct {
		AssetAverage int `gorm:"column:asset_average"`
	}
	if err := configs.DB.Table("daily_assets").Select("asset_average").Where("branch_id = ? AND asset_date BETWEEN ? AND ?", branchID, startDate, endDate).Order("asset_date DESC").Limit(1).Scan(&latestAvg).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get assets failed", err)
	}
	var formattedDailyAsset []models.DetailDailyAsset
	for _, daily := range dailyAssetFromDB {
		formattedDailyAsset = append(formattedDailyAsset, models.DetailDailyAsset{ID: daily.ID, AssetDate: helpers.FormatIndonesianDate(daily.AssetDate), AssetValue: daily.AssetValue, BranchId: daily.BranchId, AssetAverage: daily.AssetAverage, BranchName: daily.BranchName})
	}
	totalPages := services.CalculateDailyAssetTotalPages(total, limit)
	return h.jsonResponseGetAllAssets(c, fiber.StatusOK, "Daily assets retrieved successfully", latestAvg.AssetAverage, int(total), page, totalPages, limit, formattedDailyAsset)
}

func (h *DailyAssetHandler) jsonResponseGetAllAssets(c *fiber.Ctx, status int, message string, monthlyAssetAverage int, total int, page int, totalPages int, limit int, data interface{}) error {
	resp := map[string]interface{}{"monthly_asset_average": monthlyAssetAverage, "total": total, "page": page, "total_pages": totalPages, "limit": limit, "data": data}
	return helpers.JSONResponse(c, status, message, resp)
}
