package excels

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type ExcelLeastSellingHandler struct {
	excelService *export_services.ExportServices
}

func NewExcelLeastSellingHandler(excelService *export_services.ExportServices) *ExcelLeastSellingHandler {
	return &ExcelLeastSellingHandler{excelService: excelService}
}

func (h *ExcelLeastSellingHandler) ExportExcel(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)

	excelBytes, err := h.excelService.ExportLeastSellingToExcel(branchID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate excel: %v", err),
		})
	}

	timestamp := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("least-selling-%s.xlsx", timestamp)

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(excelBytes)
}
