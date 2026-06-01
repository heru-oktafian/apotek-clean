package excels

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type ExcelPurchaseItemHandler struct {
	excelService *export_services.ExportServices
}

func NewExcelPurchaseItemHandler(excelService *export_services.ExportServices) *ExcelPurchaseItemHandler {
	return &ExcelPurchaseItemHandler{excelService: excelService}
}

func (h *ExcelPurchaseItemHandler) ExportExcel(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	purchaseID := c.Query("purchase_id", "")

	if purchaseID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "purchase_id parameter is required",
		})
	}

	excelBytes, err := h.excelService.ExportPurchaseItemsToExcel(branchID, purchaseID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate excel: %v", err),
		})
	}

	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("DETAIL-PEMBELIAN-%s-%s.xlsx", purchaseID, timestamp)

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(excelBytes)
}
