package pdfs

import (
	"fmt"
	"strconv"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	services "github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type PdfProductLabelHandler struct {
	pdfService *export_services.ExportServices
}

func NewPdfProductLabelHandler(pdfService *export_services.ExportServices) *PdfProductLabelHandler {
	return &PdfProductLabelHandler{pdfService: pdfService}
}

func (h *PdfProductLabelHandler) ExportPDF(c *fiber.Ctx) error {
	productID := c.Params("id")
	qtyStr := c.Query("qty", "1")
	qty, err := strconv.Atoi(qtyStr)
	if err != nil {
		qty = 1
	}

	// Ambil branch_id dari JWT middleware
	branchID, _ := services.GetBranchID(c)

	pdfBytes, err := h.pdfService.ExportProductLabelToPDF(productID, branchID, qty)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate label: %v", err),
		})
	}

	// Generate filename dengan timestamp
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("label-%s-%s.pdf", productID, timestamp)

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(pdfBytes)
}
