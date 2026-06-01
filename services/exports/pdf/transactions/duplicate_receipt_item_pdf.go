package transactions

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	services "github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type PdfDuplicateReceiptItemHandler struct {
	pdfService *export_services.ExportServices
}

func NewPdfDuplicateReceiptItemHandler(service *export_services.ExportServices) *PdfDuplicateReceiptItemHandler {
	return &PdfDuplicateReceiptItemHandler{pdfService: service}
}

func (h *PdfDuplicateReceiptItemHandler) ExportPDF(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	duplicateReceiptID := c.Query("duplicate_receipt_id", "")

	if duplicateReceiptID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "duplicate_receipt_id parameter is required",
		})
	}

	pdfBytes, err := h.pdfService.ExportDuplicateReceiptItemsToPDF(branchID, duplicateReceiptID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate pdf: %v", err),
		})
	}

	// Generate filename
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("DETAIL-KOPI-RESEP-%s-%s.pdf", duplicateReceiptID, timestamp)

	// Set headers for download
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(pdfBytes)
}
