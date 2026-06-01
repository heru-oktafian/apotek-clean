package excels

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type ExcelBuyReturnItemHandler struct {
	excelService *export_services.ExportServices
}

func NewExcelBuyReturnItemHandler(excelService *export_services.ExportServices) *ExcelBuyReturnItemHandler {
	return &ExcelBuyReturnItemHandler{excelService: excelService}
}

func (h *ExcelBuyReturnItemHandler) ExportExcel(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	buyReturnID := c.Query("buy_return_id", "")

	if buyReturnID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "buy_return_id parameter is required",
		})
	}

	excelBytes, err := h.excelService.ExportBuyReturnItemsToExcel(branchID, buyReturnID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate excel: %v", err),
		})
	}

	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("DETAIL-RETUR-PEMBELIAN-%s-%s.xlsx", buyReturnID, timestamp)

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(excelBytes)
}
