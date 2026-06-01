package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/heru-oktafian/fiber-apotek/models"
	"github.com/xuri/excelize/v2"
)

func (s *ExportServices) ExportSalesToExcel(branchID string, month string) ([]byte, error) {
	var sales []models.Sales

	query := s.db.Where("branch_id = ?", branchID)

	if month != "" {
		parsedTime, err := time.Parse("2006-01", month)
		if err == nil {
			startDate := parsedTime
			endDate := parsedTime.AddDate(0, 1, 0)
			query = query.Where("sale_date >= ? AND sale_date < ?", startDate, endDate)
		}
	}

	err := query.Order("sale_date DESC").Find(&sales).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sales: %w", err)
	}

	f := excelize.NewFile()
	sheet := "Sales"
	f.SetSheetName("Sheet1", sheet)

	f.SetCellValue(sheet, "A1", "PENJUALAN "+month)
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	f.SetCellStyle(sheet, "A1", "F1", titleStyle)
	f.SetRowHeight(sheet, 1, 25)

	headers := []string{"ID", "KETERANGAN", "TANGGAL", "PEMBAYARAN", "SALES", "MARGIN"}
	for i, h := range headers {
		cell, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet, cell+"3", h)
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	f.SetCellStyle(sheet, "A3", "F3", headerStyle)

	styleCenter, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	styleLeft, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	styleRight, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})

	var grandTotal int
	var grandMargin int
	for i, sale := range sales {
		row := i + 4

		// Logika KETERANGAN: Ambil nama item (seperti di GetAllSalesDetail)
		var itemNames []string
		if err := s.db.Table("sale_items sit").
			Select("pro.name").
			Joins("LEFT JOIN products pro ON pro.id = sit.product_id").
			Where("sit.sale_id = ?", sale.ID).
			Order("pro.name ASC").
			Pluck("pro.name", &itemNames).Error; err != nil {
			log.Printf("[ExportSalesToExcel] Failed to fetch item names for %s: %v", sale.ID, err)
		}

		descItems := strings.Join(itemNames, ", ")
		dateWith7 := sale.SaleDate.Add(7 * time.Hour).Format("02-01-2006 15:04")
		var description string
		if descItems != "" {
			description = descItems + " ; " + dateWith7
		} else {
			description = dateWith7
		}

		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), sale.ID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), description)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), sale.SaleDate.Format("02/01/2006"))
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), string(sale.Payment))
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), formatRupiah(sale.TotalSale))
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), formatRupiah(sale.ProfitEstimate))
		grandTotal += sale.TotalSale
		grandMargin += sale.ProfitEstimate

		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styleLeft)
		f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), styleRight)
		f.SetCellStyle(sheet, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), styleRight)
	}

	totalRow := len(sales) + 4
	f.SetCellValue(sheet, fmt.Sprintf("A%d", totalRow), "GRAND TOTAL")
	f.MergeCell(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("D%d", totalRow))
	f.SetCellValue(sheet, fmt.Sprintf("E%d", totalRow), formatRupiah(grandTotal))
	f.SetCellValue(sheet, fmt.Sprintf("F%d", totalRow), formatRupiah(grandMargin))

	grandTotalStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF", Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("E%d", totalRow), grandTotalStyle)
	f.SetCellStyle(sheet, fmt.Sprintf("F%d", totalRow), fmt.Sprintf("F%d", totalRow), grandTotalStyle)
	f.SetRowHeight(sheet, totalRow, 20)

	f.SetColWidth(sheet, "A", "A", 18)
	f.SetColWidth(sheet, "B", "B", 40)
	f.SetColWidth(sheet, "C", "C", 15)
	f.SetColWidth(sheet, "D", "D", 18)
	f.SetColWidth(sheet, "E", "E", 18)
	f.SetColWidth(sheet, "F", "F", 18)

	tableErr := f.AddTable(sheet, &excelize.Table{
		Range:             fmt.Sprintf("A3:F%d", len(sales)+3),
		Name:              "SalesTable",
		StyleName:         "TableStyleMedium9",
		ShowFirstColumn:   false,
		ShowLastColumn:    false,
		ShowColumnStripes: false,
	})
	if tableErr != nil {
		log.Printf("[ExportSalesToExcel] AddTable warning: %v", tableErr)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write excel: %w", err)
	}

	return buf.Bytes(), nil
}
