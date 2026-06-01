package excels

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type ExcelNearedReportHandler struct {
	excelService *export_services.ExportServices
}

func NewExcelNearedReportHandler(excelService *export_services.ExportServices) *ExcelNearedReportHandler {
	return &ExcelNearedReportHandler{excelService: excelService}
}

func (h *ExcelNearedReportHandler) ExportExcel(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)

	excelBytes, err := h.excelService.ExportNearedReportToExcel(branchID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate excel: %v", err),
		})
	}

	// Gunakan format tanggal YYYY-MM-DD untuk nama file agar konsisten dan mudah dibaca
	timestamp := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("neared-report-%s.xlsx", timestamp)

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(excelBytes)
}
