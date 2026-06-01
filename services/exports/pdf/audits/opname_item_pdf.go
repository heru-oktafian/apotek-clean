package pdfs

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	services "github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type PdfOpnameItemHandler struct {
	pdfService *export_services.ExportServices
}

func NewPdfOpnameItemHandler(pdfService *export_services.ExportServices) *PdfOpnameItemHandler {
	return &PdfOpnameItemHandler{pdfService: pdfService}
}

func (h *PdfOpnameItemHandler) ExportPDF(c *fiber.Ctx) error {
	// Ambil branch_id dari JWT middleware
	branchID, _ := services.GetBranchID(c)

	// Ambil opname_id parameter dari query string
	opnameID := c.Query("opname_id", "")

	if opnameID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "opname_id parameter is required",
		})
	}

	pdfBytes, err := h.pdfService.ExportOpnameItemsToPDF(branchID, opnameID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate pdf: %v", err),
		})
	}

	// Generate filename dengan timestamp: opname-items-{ID}-YYYY-MM-DD-HH-MM-SS.pdf
	timestamp := time.Now().Format("2006-01-02-15:04:05")
	filename := fmt.Sprintf("DETAIL-OPNAME-%s.pdf", timestamp)

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(pdfBytes)
}
