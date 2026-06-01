package pdfs

import (
	"fmt"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	services "github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type PdfSupplierHandler struct {
	pdfService *export_services.ExportServices
}

func NewPdfSupplierHandler(pdfService *export_services.ExportServices) *PdfSupplierHandler {
	return &PdfSupplierHandler{pdfService: pdfService}
}

func (h *PdfSupplierHandler) ExportPDF(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)

	pdfBytes, err := h.pdfService.ExportSuppliersToPDF(branchID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate pdf: %v", err),
		})
	}

	timestamp := time.Now().Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("suppliers-%s.pdf", timestamp)

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(pdfBytes)
}
