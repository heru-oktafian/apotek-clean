package services

import (
	"fmt"
	"log"

	"github.com/heru-oktafian/fiber-apotek/models"
	"github.com/xuri/excelize/v2"
)

// ExportUnitsToExcel menghasilkan file Excel dengan data units per branch
func (s *ExportServices) ExportUnitsToExcel(branchID string) ([]byte, error) {
	var units []models.Unit

	err := s.db.Where("branch_id = ?", branchID).Order("name ASC").Find(&units).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch units: %w", err)
	}

	f := excelize.NewFile()
	sheet := "Units"
	f.SetSheetName("Sheet1", sheet)

	// === ROW 1: JUDUL ===
	f.SetCellValue(sheet, "A1", "DATA UNITS")
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1565C0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	f.SetCellStyle(sheet, "A1", "B1", titleStyle)
	f.SetRowHeight(sheet, 1, 25)

	// === ROW 2: JARAK (kosong) ===
	// (tidak perlu action, biarkan kosong)

	// === ROW 3: HEADER ===
	headers := []string{"UNIT ID", "UNIT NAME"}
	for i, h := range headers {
		cell, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet, cell+"3", h)
	}

	// Style Header
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
	f.SetCellStyle(sheet, "A3", "B3", headerStyle)

	// === ROW 4+: DATA ===
	for i, unit := range units {
		row := i + 4
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), unit.ID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), unit.Name)
	}

	f.SetColWidth(sheet, "A", "A", 20)
	f.SetColWidth(sheet, "B", "B", 30)

	// Buat Table
	tableErr := f.AddTable(sheet, &excelize.Table{
		Range:             fmt.Sprintf("A3:B%d", len(units)+3),
		Name:              "UnitsTable",
		StyleName:         "TableStyleMedium9",
		ShowFirstColumn:   false,
		ShowLastColumn:    false,
		ShowColumnStripes: false,
	})
	if tableErr != nil {
		log.Printf("[ExportUnitsToExcel] AddTable warning: %v", tableErr)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write excel: %w", err)
	}

	return buf.Bytes(), nil
}
