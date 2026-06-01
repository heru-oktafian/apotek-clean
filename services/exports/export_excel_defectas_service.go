package services

import (
	"fmt"
	"log"
	"time"

	"github.com/heru-oktafian/fiber-apotek/models"
	"github.com/xuri/excelize/v2"
)

func (s *ExportServices) ExportDefectasToExcel(branchID string, month string) ([]byte, error) {
	var defectas []models.Defectas

	query := s.db.Where("branch_id = ?", branchID)

	if month != "" {
		parsedTime, err := time.Parse("2006-01", month)
		if err == nil {
			startDate := parsedTime
			endDate := parsedTime.AddDate(0, 1, 0)
			query = query.Where("defecta_date >= ? AND defecta_date < ?", startDate, endDate)
		}
	}

	err := query.Order("defecta_date DESC").Find(&defectas).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch defectas: %w", err)
	}

	f := excelize.NewFile()
	sheet := "Defectas"
	f.SetSheetName("Sheet1", sheet)

	// === ROW 1: JUDUL ===
	f.SetCellValue(sheet, "A1", "LAPORAN DEFECTA "+month)
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	f.SetCellStyle(sheet, "A1", "D1", titleStyle)
	f.SetRowHeight(sheet, 1, 25)

	// === ROW 3: HEADER ===
	headers := []string{"ID", "TANGGAL", "STATUS", "ESTIMASI TOTAL"}
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
	f.SetCellStyle(sheet, "A3", "D3", headerStyle)

	styleCenter, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	styleRight, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})

	// === ROW 4+: DATA ===
	var totalEstimate int
	for i, d := range defectas {
		row := i + 4
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), d.ID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), d.DefectaDate.Format("02/01/2006"))
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), string(d.DefectaStatus))
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), formatRupiah(d.TotalEstimate))
		totalEstimate += d.TotalEstimate

		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), styleRight)
	}

	// === BARIS GRAND TOTAL ===
	totalRow := len(defectas) + 4
	f.SetCellValue(sheet, fmt.Sprintf("A%d", totalRow), "GRAND TOTAL")
	f.MergeCell(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("C%d", totalRow))
	f.SetCellValue(sheet, fmt.Sprintf("D%d", totalRow), formatRupiah(totalEstimate))

	grandTotalStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF", Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("D%d", totalRow), grandTotalStyle)
	f.SetRowHeight(sheet, totalRow, 20)

	f.SetColWidth(sheet, "A", "A", 18)
	f.SetColWidth(sheet, "B", "B", 15)
	f.SetColWidth(sheet, "C", "C", 15)
	f.SetColWidth(sheet, "D", "D", 20)

	tableErr := f.AddTable(sheet, &excelize.Table{
		Range:             fmt.Sprintf("A3:D%d", len(defectas)+3),
		Name:              "DefectasTable",
		StyleName:         "TableStyleMedium9",
		ShowFirstColumn:   false,
		ShowLastColumn:    false,
		ShowColumnStripes: false,
	})
	if tableErr != nil {
		log.Printf("[ExportDefectasToExcel] AddTable warning: %v", tableErr)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write excel: %w", err)
	}

	return buf.Bytes(), nil
}
