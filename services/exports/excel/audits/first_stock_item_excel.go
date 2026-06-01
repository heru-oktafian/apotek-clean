package excels

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type ExcelFirstStockItemHandler struct {
	excelService *export_services.ExportServices
}

func NewExcelFirstStockItemHandler(excelService *export_services.ExportServices) *ExcelFirstStockItemHandler {
	return &ExcelFirstStockItemHandler{excelService: excelService}
}

func (h *ExcelFirstStockItemHandler) ExportExcel(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	firstStockID := c.Query("first_stock_id", "")

	if firstStockID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "first_stock_id parameter is required",
		})
	}

	excelBytes, err := h.excelService.ExportFirstStockItemsToExcel(branchID, firstStockID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate excel: %v", err),
		})
	}

	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("DETAIL-STOK-AWAL-%s-%s.xlsx", firstStockID, timestamp)

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(excelBytes)
}
