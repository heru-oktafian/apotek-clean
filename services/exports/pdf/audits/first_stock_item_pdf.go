package pdfs

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	services "github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type PdfFirstStockItemHandler struct {
	pdfService *export_services.ExportServices
}

func NewPdfFirstStockItemHandler(pdfService *export_services.ExportServices) *PdfFirstStockItemHandler {
	return &PdfFirstStockItemHandler{pdfService: pdfService}
}

func (h *PdfFirstStockItemHandler) ExportPDF(c *fiber.Ctx) error {
	// Ambil branch_id dari JWT middleware
	branchID, _ := services.GetBranchID(c)

	// Ambil first_stock_id parameter dari query string
	firstStockID := c.Query("first_stock_id", "")

	if firstStockID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "first_stock_id parameter is required",
		})
	}

	pdfBytes, err := h.pdfService.ExportFirstStockItemsToPDF(branchID, firstStockID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate pdf: %v", err),
		})
	}

	// Generate filename dengan timestamp: first-stock-items-{ID}-YYYY-MM-DD-HH-MM-SS.pdf
	timestamp := time.Now().Format("2006-01-02-15:04:05")
	filename := fmt.Sprintf("DETAIL-STOK-AWAL-%s.pdf", timestamp)

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(pdfBytes)
}
