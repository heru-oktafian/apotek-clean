package transactions

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	services "github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type PdfPurchaseItemHandler struct {
	pdfService *export_services.ExportServices
}

func NewPdfPurchaseItemHandler(pdfService *export_services.ExportServices) *PdfPurchaseItemHandler {
	return &PdfPurchaseItemHandler{pdfService: pdfService}
}

func (h *PdfPurchaseItemHandler) ExportPDF(c *fiber.Ctx) error {
	// Ambil branch_id dari JWT middleware
	branchID, _ := services.GetBranchID(c)

	// Ambil purchase_id parameter dari query string
	purchaseID := c.Query("purchase_id", "")

	if purchaseID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "purchase_id parameter is required",
		})
	}

	pdfBytes, err := h.pdfService.ExportPurchaseItemsToPDF(branchID, purchaseID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate pdf: %v", err),
		})
	}

	// Generate filename
	filename := fmt.Sprintf("DETAIL-PEMBELIAN-%s.pdf", purchaseID)

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(pdfBytes)
}
