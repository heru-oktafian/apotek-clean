package transactions

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	services "github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type PdfSaleReturnItemHandler struct {
	pdfService *export_services.ExportServices
}

func NewPdfSaleReturnItemHandler(pdfService *export_services.ExportServices) *PdfSaleReturnItemHandler {
	return &PdfSaleReturnItemHandler{pdfService: pdfService}
}

func (h *PdfSaleReturnItemHandler) ExportPDF(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)

	saleReturnID := c.Query("sale_return_id", "")

	if saleReturnID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "sale_return_id parameter is required",
		})
	}

	pdfBytes, err := h.pdfService.ExportSaleReturnItemsToPDF(branchID, saleReturnID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate pdf: %v", err),
		})
	}

	timestamp := time.Now().Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("DETAIL-RETUR-PENJUALAN-%s-%s.pdf", saleReturnID, timestamp)

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(pdfBytes)
}
