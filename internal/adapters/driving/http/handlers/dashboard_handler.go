package handlers

import (
	http "net/http"
	time "time"

	configs "apotek-clean/configs"
	helpers "apotek-clean/helpers"
	models "apotek-clean/models"
	services "apotek-clean/services"
	fiber "github.com/gofiber/fiber/v2"
)

type DashboardHandler struct{}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

func (h *DashboardHandler) MonthlyProfitReport(c *fiber.Ctx) error {
	nowWIB := time.Now().In(configs.Location)
	db := configs.DB
	branchID, _ := services.GetBranchID(c)
	startOfMonth := time.Date(nowWIB.Year(), nowWIB.Month(), 1, 0, 0, 0, 0, nowWIB.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, -1)
	var summariesDB []models.DailySummaryDB
	var summariesResponse []models.DailySummaryResponse
	err := db.Table("daily_profit_reports").
		Select("report_date, SUM(total_sales) AS total_sales, SUM(profit_estimate) AS profit_estimate").
		Where("report_date BETWEEN ? AND ? AND branch_id = ?", startOfMonth, endOfMonth, branchID).
		Group("report_date").
		Order("report_date").
		Scan(&summariesDB).Error
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to retrieve monthly profit report", err)
	}
	for _, s := range summariesDB {
		summariesResponse = append(summariesResponse, models.DailySummaryResponse{ReportDate: s.ReportDate.Format("02"), TotalSales: s.TotalSales, ProfitEstimate: s.ProfitEstimate})
	}
	var monthSales, monthProfit int
	for _, s := range summariesResponse {
		monthSales += s.TotalSales
		monthProfit += s.ProfitEstimate
	}
	return h.jsonProfitReportMonthly(c, http.StatusOK, "Sales & Profit Report this month", monthSales, monthProfit, summariesResponse)
}

func (h *DashboardHandler) WeeklyProfitReport(c *fiber.Ctx) error {
	nowWIB := time.Now().In(configs.Location)
	db := configs.DB
	branchID, _ := services.GetBranchID(c)
	weekday := int(nowWIB.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	startOfWeek := time.Date(nowWIB.Year(), nowWIB.Month(), nowWIB.Day()-weekday+1, 0, 0, 0, 0, nowWIB.Location())
	endOfWeek := startOfWeek.AddDate(0, 0, 6)
	var summariesDB []models.DailySummaryDB
	err := db.Table("daily_profit_reports").
		Select("report_date, SUM(total_sales) as total_sales, SUM(profit_estimate) as profit_estimate").
		Where("report_date BETWEEN ? AND ? AND branch_id = ?", startOfWeek, endOfWeek, branchID).
		Group("report_date").
		Order("report_date").
		Scan(&summariesDB).Error
	if err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to retrieve weekly profit report", err)
	}
	var totalOmset, totalProfit int
	for _, s := range summariesDB {
		totalOmset += s.TotalSales
		totalProfit += s.ProfitEstimate
	}
	totalHPP, hppPercentage, profitPercentage := services.CalculateProfitPercentages(totalOmset, totalProfit)
	response := models.WeeklyProfitReportResponse{Omset: totalOmset, Profit: totalProfit, TotalHPP: totalHPP, ProfitPercentage: profitPercentage, HPPPercentage: hppPercentage}
	return helpers.JSONResponse(c, http.StatusOK, "Weekly sales & profit report", []models.WeeklyProfitReportResponse{response})
}

func (h *DashboardHandler) DailyProfitReport(c *fiber.Ctx) error {
	db := configs.DB
	branchID, _ := services.GetBranchID(c)
	today := time.Now().In(configs.Location).Format("2006-01-02")
	var summary models.DailySummaryResponse
	err := db.Table("daily_profit_reports").
		Select("report_date, SUM(total_sales) AS total_sales, SUM(profit_estimate) AS profit_estimate").
		Where("report_date = ? AND branch_id = ?", today, branchID).
		Group("report_date").
		Scan(&summary).Error
	if err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to retrieve daily profit report", err)
	}
	if summary.ReportDate == "" {
		summary = models.DailySummaryResponse{ReportDate: today, TotalSales: 0, ProfitEstimate: 0}
	}
	return helpers.JSONResponse(c, http.StatusOK, "Sales & Profit Report today", summary)
}

func (h *DashboardHandler) GetTopSellingProducts(c *fiber.Ctx) error {
	db := configs.DB
	branchID, _ := services.GetBranchID(c)
	oneMonthAgo := time.Now().AddDate(0, -1, 0)
	type Result struct {
		ProductID string
		Name      string
		TotalQty  int
	}
	var results []Result
	err := db.Table("sale_items").
		Select("products.id as product_id, products.name, SUM(sale_items.qty) as total_qty").
		Joins("JOIN sales ON sales.id = sale_items.sale_id").
		Joins("JOIN products ON products.id = sale_items.product_id").
		Where("sales.sale_date >= ? AND sales.branch_id = ?", oneMonthAgo, branchID).
		Group("products.id, products.name").
		Order("total_qty DESC").
		Limit(10).
		Scan(&results).Error
	if err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to fetch top selling products", err)
	}
	return helpers.JSONResponse(c, http.StatusOK, "Top Selling Products Last Month", results)
}

func (h *DashboardHandler) GetLeastSellingProducts(c *fiber.Ctx) error {
	db := configs.DB
	branchID, _ := services.GetBranchID(c)
	now := time.Now()
	oneMonthAgo := now.AddDate(0, -1, 0)
	subQuery := db.Table("sale_items").Select("product_id, SUM(qty) as total_sold").Joins("JOIN sales ON sales.id = sale_items.sale_id").Where("sales.sale_date BETWEEN ? AND ?", oneMonthAgo, now).Group("product_id")
	type Result struct {
		ProductID   string `json:"product_id"`
		ProductName string `json:"product_name"`
		Stock       int    `json:"stock"`
		TotalSold   int    `json:"total_sold"`
	}
	var results []Result
	if err := db.Table("products p").
		Select("p.id as product_id, p.name as product_name, p.stock, COALESCE(s.total_sold, 0) as total_sold").
		Joins("LEFT JOIN (?) as s ON p.id = s.product_id", subQuery).
		Where("p.stock >= ? AND p.branch_id = ?", 1, branchID).
		Order("total_sold ASC").
		Limit(25).
		Scan(&results).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to retrieve least selling products", err)
	}
	return helpers.JSONResponse(c, http.StatusOK, "Least selling products (1 month)", results)
}

func (h *DashboardHandler) GetExpiringProducts(c *fiber.Ctx) error {
	db := configs.DB
	nowWIB := time.Now().In(configs.Location)
	threeMonthsLater := nowWIB.AddDate(0, 3, 0)
	type ProductQueryResult struct {
		ID          string    `gorm:"column:id"`
		SKU         string    `gorm:"column:sku"`
		Name        string    `gorm:"column:name"`
		Stock       int       `gorm:"column:stock"`
		Unit        string    `gorm:"column:unit"`
		ExpiredDate time.Time `gorm:"column:expired_date"`
	}
	var rawProducts []ProductQueryResult
	err := db.Table("products").
		Select("products.id, products.sku, products.name, products.stock, units.name as unit, products.expired_date").
		Joins("LEFT JOIN units ON products.unit_id = units.id").
		Where("products.expired_date <= ? AND products.stock >= ?", threeMonthsLater, 1).
		Order("products.expired_date ASC").
		Scan(&rawProducts).Error
	if err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to fetch expiring products", err)
	}
	var productsResponse []models.ProductExpiredResponse
	for _, p := range rawProducts {
		productsResponse = append(productsResponse, models.ProductExpiredResponse{ID: p.ID, SKU: p.SKU, Name: p.Name, Stock: p.Stock, Unit: p.Unit, ExpiredDate: p.ExpiredDate.Format("2006-01-02")})
	}
	return helpers.JSONResponse(c, http.StatusOK, "All Product Near Expired (<= 3 month)", productsResponse)
}

func (h *DashboardHandler) GetDailyProfitReportByUser(c *fiber.Ctx) error {
	nowWIB := time.Now().In(configs.Location)
	db := configs.DB
	branchID, _ := services.GetBranchID(c)
	today := nowWIB.Format("2006-01-02")
	type Result struct {
		UserID   string
		UserName string
		Profit   int
		Sales    int
	}
	var results []Result
	err := db.Table("daily_profit_reports").
		Select("users.id, users.name AS user_name, SUM(daily_profit_reports.profit_estimate) AS profit, SUM(daily_profit_reports.total_sales) AS sales").
		Joins("JOIN users ON users.id = daily_profit_reports.user_id").
		Where("daily_profit_reports.report_date = ? AND daily_profit_reports.branch_id = ?", today, branchID).
		Group("users.id, users.name").
		Scan(&results).Error
	if err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to fetch report", err)
	}
	var totalProfit, totalSales int
	for _, r := range results {
		totalProfit += r.Profit
		totalSales += r.Sales
	}
	var qtyTransactions int64
	err = db.Table("sales").Where("DATE(created_at) = ? AND branch_id = ?", today, branchID).Count(&qtyTransactions).Error
	if err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Failed to count transactions", err)
	}
	abvTransactions := 0
	if qtyTransactions > 0 {
		abvTransactions = int(totalSales) / int(qtyTransactions)
	}
	rawReportData := make([]fiber.Map, 0, len(results))
	for _, r := range results {
		rawReportData = append(rawReportData, fiber.Map{"user_id": r.UserID, "user_name": r.UserName, "profit": r.Profit, "sales": r.Sales})
	}
	reportData := services.BuildDailyProfitByUserReportData(rawReportData, totalProfit)
	return h.jsonProfitReportToDay(c, http.StatusOK, "Profit Report Successfully", "daily", totalProfit, totalSales, qtyTransactions, abvTransactions, reportData)
}

func (h *DashboardHandler) jsonProfitReportToDay(c *fiber.Ctx, status int, message string, reportType string, totalProfit int, totalSales int, qtyTransactions int64, abvTransactions int, data interface{}) error {
	resp := models.ResponseProfitReportToDay{Status: http.StatusText(status), Message: message, ReportType: reportType, TotalProfit: totalProfit, TotalSales: totalSales, QtyTransactions: qtyTransactions, AbvTransactions: abvTransactions, Data: data}
	return helpers.JSONResponse(c, status, message, resp)
}

func (h *DashboardHandler) jsonProfitReportMonthly(c *fiber.Ctx, status int, message string, monthSales int, monthProfit int, data interface{}) error {
	resp := models.ResponseProfitReportMonthly{Status: http.StatusText(status), Message: message, MonthSales: monthSales, MonthProfit: monthProfit, Data: data}
	return helpers.JSONResponse(c, status, message, resp)
}
