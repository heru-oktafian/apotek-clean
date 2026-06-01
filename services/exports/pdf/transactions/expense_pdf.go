package transactions

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type PdfExpenseHandler struct {
	pdfService *export_services.ExportServices
}

func NewPdfExpenseHandler(pdfService *export_services.ExportServices) *PdfExpenseHandler {
	return &PdfExpenseHandler{pdfService: pdfService}
}

func (h *PdfExpenseHandler) ExportPDF(c *fiber.Ctx) error {
	// Ambil branch_id dari JWT middleware
	branchID, _ := services.GetBranchID(c)

	// Ambil month parameter dari query string (format: YYYY-MM)
	month := c.Query("month", "")

	pdfBytes, err := h.pdfService.ExportExpensesToPDF(branchID, month)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate pdf: %v", err),
		})
	}

	// Generate filename dengan timestamp: PENGELUARAN-YYYY-MM-DD-HH:MM:SS.pdf
	timestamp := time.Now().Format("2006-01-02-15:04:05")
	filename := fmt.Sprintf("PENGELUARAN-%s.pdf", timestamp)

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(pdfBytes)
}
