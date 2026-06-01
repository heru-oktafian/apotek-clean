package services

import (
	"fmt"
	"log"
	"time"

	"github.com/heru-oktafian/fiber-apotek/models"
	"github.com/xuri/excelize/v2"
)

func (s *ExportServices) ExportExpensesToExcel(branchID string, month string) ([]byte, error) {
	var expenses []models.Expenses

	query := s.db.Where("branch_id = ?", branchID)

	if month != "" {
		parsedTime, err := time.Parse("2006-01", month)
		if err == nil {
			startDate := parsedTime
			endDate := parsedTime.AddDate(0, 1, 0)
			query = query.Where("expense_date >= ? AND expense_date < ?", startDate, endDate)
		}
	}

	err := query.Order("expense_date DESC").Find(&expenses).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expenses: %w", err)
	}

	f := excelize.NewFile()
	sheet := "Expenses"
	f.SetSheetName("Sheet1", sheet)

	f.SetCellValue(sheet, "A1", "PENGELUARAN "+month)
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	f.SetCellStyle(sheet, "A1", "E1", titleStyle)
	f.SetRowHeight(sheet, 1, 25)

	headers := []string{"ID", "KETERANGAN", "TANGGAL", "PEMBAYARAN", "TOTAL"}
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
	f.SetCellStyle(sheet, "A3", "E3", headerStyle)

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
	for i, e := range expenses {
		row := i + 4
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), e.ID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), e.Description)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), e.ExpenseDate.Format("02/01/2006"))
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), string(e.Payment))
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), formatRupiah(e.TotalExpense))
		grandTotal += e.TotalExpense

		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styleLeft)
		f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), styleRight)
	}

	totalRow := len(expenses) + 4
	f.SetCellValue(sheet, fmt.Sprintf("A%d", totalRow), "GRAND TOTAL")
	f.MergeCell(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("D%d", totalRow))
	f.SetCellValue(sheet, fmt.Sprintf("E%d", totalRow), formatRupiah(grandTotal))

	grandTotalStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF", Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("E%d", totalRow), grandTotalStyle)
	f.SetRowHeight(sheet, totalRow, 20)

	f.SetColWidth(sheet, "A", "A", 18)
	f.SetColWidth(sheet, "B", "B", 40)
	f.SetColWidth(sheet, "C", "C", 15)
	f.SetColWidth(sheet, "D", "D", 18)
	f.SetColWidth(sheet, "E", "E", 18)

	tableErr := f.AddTable(sheet, &excelize.Table{
		Range:             fmt.Sprintf("A3:E%d", len(expenses)+3),
		Name:              "ExpensesTable",
		StyleName:         "TableStyleMedium9",
		ShowFirstColumn:   false,
		ShowLastColumn:    false,
		ShowColumnStripes: false,
	})
	if tableErr != nil {
		log.Printf("[ExportExpensesToExcel] AddTable warning: %v", tableErr)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write excel: %w", err)
	}

	return buf.Bytes(), nil
}
