package controllers

import (
	math "math"
	strconv "strconv"
	strings "strings"
	time "time"

	fiber "github.com/gofiber/fiber/v2"
	configs "github.com/heru-oktafian/fiber-apotek/configs"
	helpers "github.com/heru-oktafian/fiber-apotek/helpers"
	models "github.com/heru-oktafian/fiber-apotek/models"
	services "github.com/heru-oktafian/fiber-apotek/services"
)

// JSONResponseGetAllAssets membungkus response untuk GetAllAssets
func JSONResponseGetAllAssets(c *fiber.Ctx, status int, message string, errMsg string, monthlyAssetAverage int, total int, page int, totalPages int, limit int, data interface{}) error {
	resp := map[string]interface{}{
		"monthly_asset_average": monthlyAssetAverage,
		"total":                 total,
		"page":                  page,
		"total_pages":           totalPages,
		"limit":                 limit,
		"data":                  data,
	}
	return helpers.JSONResponse(c, status, message, resp)
}

func GetAllAssets(c *fiber.Ctx) error {
	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	// Ambil ID cabang
	branchID, _ := services.GetBranchID(c)

	// Ambil parameter page dan search dari query URL
	pageParam := c.Query("page")

	// Konversi page ke int, default ke 1 jika tidak valid
	page := 1
	if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
		page = p
	}

	limit := 10                  // Tetapkan limit ke 10 data per halaman
	offset := (page - 1) * limit // Hitung offset berdasarkan halaman dan limit

	month := strings.TrimSpace(c.Query("month"))

	// Jika month kosong, isi dengan bulan ini (format YYYY-MM)
	if month == "" {
		month = nowWIB.Format("2006-01")
	}

	// Parse month ke startDate dan endDate (always parse to use for asset_average lookup)
	parsedMonth, err := time.Parse("2006-01", month)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid month format. Month should be in format YYYY-MM", err)
	}
	startDate := parsedMonth
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)

	var dailyAssetFromDB []models.AllDailyAsset // Gunakan models.DailyAsset untuk mengambil data dari DB
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

	// Ambil asset_average terbaru pada bulan yang dipilih untuk branch ini
	var latestAvg struct {
		AssetAverage int `gorm:"column:asset_average"`
	}
	if err := configs.DB.Table("daily_assets").Select("asset_average").Where("branch_id = ? AND asset_date BETWEEN ? AND ?", branchID, startDate, endDate).Order("asset_date DESC").Limit(1).Scan(&latestAvg).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get assets failed", err)
	}

	// Buat slice baru untuk menampung data yang sudah diformat
	var formattedDailyAsset []models.DetailDailyAsset
	for _, daily := range dailyAssetFromDB {
		formattedDailyAsset = append(formattedDailyAsset, models.DetailDailyAsset{
			ID:           daily.ID,
			AssetDate:    helpers.FormatIndonesianDate(daily.AssetDate), // Format tanggal di sini
			AssetValue:   daily.AssetValue,
			BranchId:     daily.BranchId,
			AssetAverage: daily.AssetAverage,
			BranchName:   daily.BranchName,
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// Gunakan JSONResponseGetAll helper dengan data yang sudah diformat
	// Gabungkan payload: data list + asset_average bulan ini (nilai terbaru dalam bulan)
	vMonthlyAssetAverage := latestAvg.AssetAverage

	return JSONResponseGetAllAssets(
		c,
		fiber.StatusOK,
		"Daily assets retrieved successfully",
		"",
		vMonthlyAssetAverage,
		int(total),
		page,
		totalPages,
		limit,
		formattedDailyAsset,
	)
}
