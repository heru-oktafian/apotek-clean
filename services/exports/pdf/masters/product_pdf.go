package pdfs

import (
	fmt "fmt"
	time "time"

	fiber "github.com/gofiber/fiber/v2"
	services "github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type PdfProductHandler struct {
	pdfService *export_services.ExportServices
}

func NewPdfProductHandler(pdfService *export_services.ExportServices) *PdfProductHandler {
	return &PdfProductHandler{pdfService: pdfService}
}

func (h *PdfProductHandler) ExportPDF(c *fiber.Ctx) error {
	// Ambil branch_id dari JWT middleware
	branchID, _ := services.GetBranchID(c)

	pdfBytes, err := h.pdfService.ExportProductsToPDF(branchID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate pdf: %v", err),
		})
	}

	// Generate filename dengan timestamp: products-YYYY-MM-DD-HH-MM-SS.pdf
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("products-%s.pdf", timestamp)

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(pdfBytes)
}
