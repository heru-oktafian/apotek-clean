package excels

import (
	fmt "fmt"
	time "time"

	fiber "github.com/gofiber/fiber/v2"
	services "github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type ExcelMemberCategoryHandler struct {
	excelService *export_services.ExportServices
}

func NewExcelMemberCategoryHandler(excelService *export_services.ExportServices) *ExcelMemberCategoryHandler {
	return &ExcelMemberCategoryHandler{excelService: excelService}
}

func (h *ExcelMemberCategoryHandler) ExportExcel(c *fiber.Ctx) error {
	// Ambil branch_id dari JWT middleware
	branchID, _ := services.GetBranchID(c)

	excelBytes, err := h.excelService.ExportMemberCategoriesToExcel(branchID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate excel: %v", err),
		})
	}

	// Generate filename dengan timestamp: member-categories-YYYY-MM-DD-HH-MM-SS.xlsx
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("member-categories-%s.xlsx", timestamp)

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(excelBytes)
}
