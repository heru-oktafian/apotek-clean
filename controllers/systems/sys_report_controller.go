package controllers

import (
	time "time"

	fiber "github.com/gofiber/fiber/v2"
	configs "github.com/heru-oktafian/fiber-apotek/configs"
	helpers "github.com/heru-oktafian/fiber-apotek/helpers"
	models "github.com/heru-oktafian/fiber-apotek/models"
	services "github.com/heru-oktafian/fiber-apotek/services"
)

func GetNeracaSaldo(c *fiber.Ctx) error {
	db := configs.DB
	branchID, _ := services.GetBranchID(c)
	month := c.Query("month") // format: "2025-05"

	if branchID == "" {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "branch_id is required", nil)
	}

	// Konversi bulan ke rentang tanggal
	var startDate, endDate time.Time
	var err error
	if month != "" {
		startDate, err = time.Parse("2006-01", month)
		if err != nil {
			return helpers.JSONResponse(c, fiber.StatusBadRequest, "Format bulan tidak valid. Gunakan format YYYY-MM.", nil)
		}
		endDate = startDate.AddDate(0, 1, 0) // awal bulan berikutnya
	}

	type Summary struct {
		TransactionType string
		TransactionDate string
		Total           int
	}

	var summaries []Summary

	query := db.Table("transaction_reports").
		Select("transaction_type, DATE(created_at) AS transaction_date, SUM(total) AS total").
		Where("branch_id = ? AND payment != 'paid_by_credit'", branchID).
		Group("transaction_type, DATE(created_at)").
		Order("transaction_date ASC")

	if month != "" {
		query = query.Where("created_at >= ? AND created_at < ?", startDate, endDate)
	}

	if err := query.Scan(&summaries).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data transaksi", err)
	}

	// Kategorikan dan hitung
	var debit []fiber.Map
	var credit []fiber.Map
	var totalDebit, totalCredit int

	for _, s := range summaries {
		entry := fiber.Map{
			"transaction_type":  s.TransactionType,
			"transaction_date":  s.TransactionDate,
			"total_transaction": s.Total,
		}

		switch s.TransactionType {
		case string(models.Sale), string(models.Income), string(models.BuyReturn):
			debit = append(debit, entry)
			totalDebit += s.Total
		case string(models.Purchase), string(models.Expense), string(models.SaleReturn):
			credit = append(credit, entry)
			totalCredit += s.Total
		}
	}

	totalSaldo := totalDebit - totalCredit

	return helpers.JSONResponse(c, fiber.StatusOK, "Success", fiber.Map{
		"debit":        debit,
		"credit":       credit,
		"total_debit":  totalDebit,
		"total_credit": totalCredit,
		"total_saldo":  totalSaldo,
	})
}

// GetProfitGraphByMonth get profit graph by selected month
func GetProfitGraphByMonth(c *fiber.Ctx) error {

	db := configs.DB
	branchID, _ := services.GetBranchID(c)
	month := c.Query("month") // format: YYYY-MM

	parsedMonth, err := time.Parse("2006-01", month)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid month format. Use YYYY-MM.", nil)
	}

	startOfMonth := time.Date(parsedMonth.Year(), parsedMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, -1)

	var summaries []models.DailySummaryDB

	err = db.Table("daily_profit_reports").
		Select("report_date, SUM(total_sales) AS total_sales, SUM(profit_estimate) AS profit_estimate").
		Where("report_date BETWEEN ? AND ? AND branch_id = ?", startOfMonth, endOfMonth, branchID).
		Group("report_date").
		Order("report_date").
		Scan(&summaries).Error

	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data laporan", err)
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Sales & Profit Report on "+month, summaries)
}
