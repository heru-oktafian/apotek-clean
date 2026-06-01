package excels

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type ExcelSaleItemHandler struct {
	excelService *export_services.ExportServices
}

func NewExcelSaleItemHandler(excelService *export_services.ExportServices) *ExcelSaleItemHandler {
	return &ExcelSaleItemHandler{excelService: excelService}
}

func (h *ExcelSaleItemHandler) ExportExcel(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	saleID := c.Query("sale_id", "")

	if saleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "sale_id parameter is required",
		})
	}

	excelBytes, err := h.excelService.ExportSaleItemsToExcel(branchID, saleID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate excel: %v", err),
		})
	}

	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("DETAIL-PENJUALAN-%s-%s.xlsx", saleID, timestamp)

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(excelBytes)
}
