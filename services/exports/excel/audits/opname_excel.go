package excels

import (
	fmt "fmt"
	time "time"

	fiber "github.com/gofiber/fiber/v2"
	services "github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type ExcelOpnameHandler struct {
	excelService *export_services.ExportServices
}

func NewExcelOpnameHandler(excelService *export_services.ExportServices) *ExcelOpnameHandler {
	return &ExcelOpnameHandler{excelService: excelService}
}

func (h *ExcelOpnameHandler) ExportExcel(c *fiber.Ctx) error {
	// Ambil branch_id dari JWT middleware
	branchID, _ := services.GetBranchID(c)

	// Ambil month parameter dari query string (format: YYYY-MM)
	month := c.Query("month", "")

	excelBytes, err := h.excelService.ExportOpnamesToExcel(branchID, month)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate excel: %v", err),
		})
	}

	// Generate filename dengan timestamp: opnames-YYYY-MM-DD-HH-MM-SS.xlsx
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("opnames-%s.xlsx", timestamp)

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(excelBytes)
}
