package excels

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type ExcelOpnameItemHandler struct {
	excelService *export_services.ExportServices
}

func NewExcelOpnameItemHandler(excelService *export_services.ExportServices) *ExcelOpnameItemHandler {
	return &ExcelOpnameItemHandler{excelService: excelService}
}

func (h *ExcelOpnameItemHandler) ExportExcel(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	opnameID := c.Query("opname_id", "")

	if opnameID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "opname_id parameter is required",
		})
	}

	excelBytes, err := h.excelService.ExportOpnameItemsToExcel(branchID, opnameID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate excel: %v", err),
		})
	}

	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("DETAIL-OPNAME-%s-%s.xlsx", opnameID, timestamp)

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(excelBytes)
}
