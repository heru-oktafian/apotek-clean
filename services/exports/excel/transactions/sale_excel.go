// Package excels — HTTP handler layer for Excel exports.
// This is Layer 2 (HTTP layer) of the export architecture.
// Layer 1 (services/exports/export_excel_*.go) does the actual Excel file generation.
// This layer handles: extracting branch_id from JWT, setting Content-Type/Disposition, returning bytes.
// See: internal/adapters/driving/http/routes/export_routes.go for full architecture overview.

package excels

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"apotek-clean/services"
	export_services "apotek-clean/services/exports"
)

type ExcelSaleHandler struct {
	excelService *export_services.ExportServices
}

func NewExcelSaleHandler(excelService *export_services.ExportServices) *ExcelSaleHandler {
	return &ExcelSaleHandler{excelService: excelService}
}

func (h *ExcelSaleHandler) ExportExcel(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	month := c.Query("month", "")

	excelBytes, err := h.excelService.ExportSalesToExcel(branchID, month)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate excel: %v", err),
		})
	}

	timestamp := time.Now().Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("sales-%s.xlsx", timestamp)

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(excelBytes)
}
