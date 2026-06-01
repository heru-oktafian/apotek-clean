package excels

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/heru-oktafian/fiber-apotek/services"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
)

type ExcelNeracaSaldoHandler struct {
	excelService *export_services.ExportServices
}

func NewExcelNeracaSaldoHandler(excelService *export_services.ExportServices) *ExcelNeracaSaldoHandler {
	return &ExcelNeracaSaldoHandler{excelService: excelService}
}

func (h *ExcelNeracaSaldoHandler) ExportExcel(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	month := c.Query("month", "")

	excelBytes, err := h.excelService.ExportNeracaSaldoToExcel(branchID, month)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("gagal generate excel: %v", err),
		})
	}

	timestamp := time.Now().Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("neraca-saldo-%s.xlsx", timestamp)

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(excelBytes)
}
