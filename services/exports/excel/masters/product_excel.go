package excels

import (
	fmt "fmt"
	log "log"

	fiber "github.com/gofiber/fiber/v2"
	services "github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type ExcelProductHandler struct {
	excelService *export_services.ExportServices
}

func NewExcelProductHandler(excelService *export_services.ExportServices) *ExcelProductHandler {
	return &ExcelProductHandler{excelService: excelService}
}

// ExportExcel → Export produk ke file Excel
func (h *ExcelProductHandler) ExportExcel(c *fiber.Ctx) error {
	// Ambil branch_id dari JWT middleware
	branchID, _ := services.GetBranchID(c)

	if branchID == "" {
		log.Println("[ExportExcel] ERROR: branch_id kosong")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "branch_id tidak ditemukan",
		})
	}

	excelBytes, err := h.excelService.ExportProductsToExcel(branchID)
	if err != nil {
		log.Printf("[ExportExcel] ERROR: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate excel: %v", err),
		})
	}

	if len(excelBytes) == 0 {
		log.Println("[ExportExcel] ERROR: Excel bytes kosong")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "gagal generate file excel - bytes kosong",
		})
	}

	// Set response headers untuk file download
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", `attachment; filename="produk.xlsx"`)
	c.Set("Content-Length", fmt.Sprintf("%d", len(excelBytes)))

	return c.Send(excelBytes)
}
