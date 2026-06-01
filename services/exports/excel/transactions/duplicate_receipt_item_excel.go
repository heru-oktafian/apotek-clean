package excels

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type ExcelDuplicateReceiptItemHandler struct {
	excelService *export_services.ExportServices
}

func NewExcelDuplicateReceiptItemHandler(excelService *export_services.ExportServices) *ExcelDuplicateReceiptItemHandler {
	return &ExcelDuplicateReceiptItemHandler{excelService: excelService}
}

func (h *ExcelDuplicateReceiptItemHandler) ExportExcel(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	duplicateReceiptID := c.Query("duplicate_receipt_id", "")

	if duplicateReceiptID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "duplicate_receipt_id parameter is required",
		})
	}

	excelBytes, err := h.excelService.ExportDuplicateReceiptItemsToExcel(branchID, duplicateReceiptID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate excel: %v", err),
		})
	}

	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("DETAIL-KOPI-RESEP-%s-%s.xlsx", duplicateReceiptID, timestamp)

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(excelBytes)
}
