package transactions

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	services "github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type PdfSaleItemHandler struct {
	pdfService *export_services.ExportServices
}

func NewPdfSaleItemHandler(pdfService *export_services.ExportServices) *PdfSaleItemHandler {
	return &PdfSaleItemHandler{pdfService: pdfService}
}

func (h *PdfSaleItemHandler) ExportPDF(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)

	saleID := c.Query("sale_id", "")

	if saleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "sale_id parameter is required",
		})
	}

	pdfBytes, err := h.pdfService.ExportSaleItemsToPDF(branchID, saleID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate pdf: %v", err),
		})
	}

	timestamp := time.Now().Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("DETAIL-PENJUALAN-%s-%s.pdf", saleID, timestamp)

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(pdfBytes)
}
