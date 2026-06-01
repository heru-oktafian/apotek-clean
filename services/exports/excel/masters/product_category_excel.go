package excels

import (
	fmt "fmt"
	time "time"

	fiber "github.com/gofiber/fiber/v2"
	services "github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type ExcelProductCategoryHandler struct {
	excelService *export_services.ExportServices
}

func NewExcelProductCategoryHandler(excelService *export_services.ExportServices) *ExcelProductCategoryHandler {
	return &ExcelProductCategoryHandler{excelService: excelService}
}

func (h *ExcelProductCategoryHandler) ExportExcel(c *fiber.Ctx) error {
	// Ambil branch_id dari JWT middleware
	branchID, _ := services.GetBranchID(c)

	excelBytes, err := h.excelService.ExportProductCategoriesToExcel(branchID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate excel: %v", err),
		})
	}

	// Generate filename dengan timestamp: product-categories-YYYY-MM-DD-HH-MM-SS.xlsx
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("product-categories-%s.xlsx", timestamp)

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(excelBytes)
}
