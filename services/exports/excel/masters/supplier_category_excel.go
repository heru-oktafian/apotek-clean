package excels

import (
	fmt "fmt"
	time "time"

	fiber "github.com/gofiber/fiber/v2"
	services "github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type ExcelSupplierCategoryHandler struct {
	excelService *export_services.ExportServices
}

func NewExcelSupplierCategoryHandler(excelService *export_services.ExportServices) *ExcelSupplierCategoryHandler {
	return &ExcelSupplierCategoryHandler{excelService: excelService}
}

func (h *ExcelSupplierCategoryHandler) ExportExcel(c *fiber.Ctx) error {
	// Ambil branch_id dari JWT middleware
	branchID, _ := services.GetBranchID(c)

	excelBytes, err := h.excelService.ExportSupplierCategoriesToExcel(branchID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate excel: %v", err),
		})
	}

	// Generate filename dengan timestamp: supplier-categories-YYYY-MM-DD-HH-MM-SS.xlsx
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("supplier-categories-%s.xlsx", timestamp)

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(excelBytes)
}
