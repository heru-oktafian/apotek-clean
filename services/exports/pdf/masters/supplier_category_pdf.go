package pdfs

import (
	fmt "fmt"
	time "time"

	fiber "github.com/gofiber/fiber/v2"
	services "github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type PdfSupplierCategoryHandler struct {
	pdfService *export_services.ExportServices
}

func NewPdfSupplierCategoryHandler(pdfService *export_services.ExportServices) *PdfSupplierCategoryHandler {
	return &PdfSupplierCategoryHandler{pdfService: pdfService}
}

func (h *PdfSupplierCategoryHandler) ExportPDF(c *fiber.Ctx) error {
	// Ambil branch_id dari JWT middleware
	branchID, _ := services.GetBranchID(c)

	pdfBytes, err := h.pdfService.ExportSupplierCategoriesToPDF(branchID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate pdf: %v", err),
		})
	}

	// Generate filename dengan timestamp: supplier-categories-YYYY-MM-DD-HH-MM-SS.pdf
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("supplier-categories-%s.pdf", timestamp)

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(pdfBytes)
}
