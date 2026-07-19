package handlers

import (
	"strconv"
	"time"

	"apotek-clean/configs"
	"apotek-clean/helpers"
	"apotek-clean/services"

	"github.com/gofiber/fiber/v2"
)

type StockTraceHandler struct{}

func NewStockTraceHandler() *StockTraceHandler {
	return &StockTraceHandler{}
}

func (h *StockTraceHandler) GetProductTrace(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	productID := c.Query("product_id")
	if productID == "" {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "product_id is required", nil)
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	traces, total, err := services.GetProductTrace(configs.DB, productID, branchID, page, limit)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to get product traces", err.Error())
	}

	// Attach branch_name
	for i := range traces {
		var branchName string
		configs.DB.Table("branches").
			Select("branch_name").
			Where("id = ?", traces[i].BranchID).
			Scan(&branchName)
		traces[i].BranchName = branchName
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Product traces retrieved successfully", "", int(total), page, totalPages, limit, traces)
}

func (h *StockTraceHandler) GetProductTraceByDateRange(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	productID := c.Query("product_id")
	if productID == "" {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "product_id is required", nil)
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	if startDateStr == "" || endDateStr == "" {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "start_date and end_date are required", nil)
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "invalid start_date format, use YYYY-MM-DD", nil)
	}
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "invalid end_date format, use YYYY-MM-DD", nil)
	}
	endDate = endDate.Add(24*time.Hour - time.Second)

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	traces, total, err := services.GetProductTraceByDateRange(configs.DB, productID, branchID, startDate, endDate, page, limit)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to get product traces", err.Error())
	}

	// Attach branch_name
	for i := range traces {
		var branchName string
		configs.DB.Table("branches").
			Select("branch_name").
			Where("id = ?", traces[i].BranchID).
			Scan(&branchName)
		traces[i].BranchName = branchName
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Product traces retrieved successfully", "", int(total), page, totalPages, limit, traces)
}
