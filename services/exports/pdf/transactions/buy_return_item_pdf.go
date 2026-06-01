package transactions

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	services "github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type PdfBuyReturnItemHandler struct {
	pdfService *export_services.ExportServices
}

func NewPdfBuyReturnItemHandler(pdfService *export_services.ExportServices) *PdfBuyReturnItemHandler {
	return &PdfBuyReturnItemHandler{pdfService: pdfService}
}

func (h *PdfBuyReturnItemHandler) ExportPDF(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)

	buyReturnID := c.Query("buy_return_id", "")

	if buyReturnID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "buy_return_id parameter is required",
		})
	}

	pdfBytes, err := h.pdfService.ExportBuyReturnItemsToPDF(branchID, buyReturnID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate pdf: %v", err),
		})
	}

	timestamp := time.Now().Format("2006-01-02-15:04:05")
	filename := fmt.Sprintf("DETAIL-RETUR-PEMBELIAN-%s.pdf", timestamp)

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(pdfBytes)
}
