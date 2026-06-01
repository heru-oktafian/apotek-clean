package transactions

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type PdfSaleHandler struct {
	pdfService *export_services.ExportServices
}

func NewPdfSaleHandler(pdfService *export_services.ExportServices) *PdfSaleHandler {
	return &PdfSaleHandler{pdfService: pdfService}
}

func (h *PdfSaleHandler) ExportPDF(c *fiber.Ctx) error {
	// Ambil branch_id dari JWT middleware
	branchID, _ := services.GetBranchID(c)

	// Ambil month parameter dari query string (format: YYYY-MM)
	month := c.Query("month", "")

	pdfBytes, err := h.pdfService.ExportSalesToPDF(branchID, month)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate pdf: %v", err),
		})
	}

	// Generate filename dengan timestamp: PENJUALAN-YYYY-MM-DD-HH:MM:SS.pdf
	timestamp := time.Now().Format("2006-01-02-15:04:05")
	filename := fmt.Sprintf("PENJUALAN-%s.pdf", timestamp)

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(pdfBytes)
}
