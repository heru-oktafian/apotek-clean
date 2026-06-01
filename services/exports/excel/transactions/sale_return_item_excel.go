package excels

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type ExcelSaleReturnItemHandler struct {
	excelService *export_services.ExportServices
}

func NewExcelSaleReturnItemHandler(excelService *export_services.ExportServices) *ExcelSaleReturnItemHandler {
	return &ExcelSaleReturnItemHandler{excelService: excelService}
}

func (h *ExcelSaleReturnItemHandler) ExportExcel(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	saleReturnID := c.Query("sale_return_id", "")

	if saleReturnID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "sale_return_id parameter is required",
		})
	}

	excelBytes, err := h.excelService.ExportSaleReturnItemsToExcel(branchID, saleReturnID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate excel: %v", err),
		})
	}

	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("DETAIL-RETUR-PENJUALAN-%s-%s.xlsx", saleReturnID, timestamp)

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(excelBytes)
}
