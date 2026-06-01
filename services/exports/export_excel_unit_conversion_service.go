package services

import (
	"fmt"
	"log"

	"github.com/heru-oktafian/fiber-apotek/models"
	"github.com/xuri/excelize/v2"
)

func (s *ExportServices) ExportUnitConversionsToExcel(branchID string) ([]byte, error) {
	var conversions []models.UnitConversion

	err := s.db.Where("branch_id = ?", branchID).Order("id ASC").Find(&conversions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch unit conversions: %w", err)
	}

	f := excelize.NewFile()
	sheet := "Unit Conversions"
	f.SetSheetName("Sheet1", sheet)

	// === ROW 1: JUDUL ===
	f.SetCellValue(sheet, "A1", "DATA UNIT CONVERSIONS")
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1565C0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	f.SetCellStyle(sheet, "A1", "E1", titleStyle)
	f.SetRowHeight(sheet, 1, 25)

	// === ROW 2: JARAK (kosong) ===
	// (tidak perlu action, biarkan kosong)

	// === ROW 3: HEADER ===
	headers := []string{"CONVERSION ID", "PRODUCT ID", "INIT UNIT", "FINAL UNIT", "VALUE CONVERSION"}
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
	f.SetCellStyle(sheet, "A3", "E3", headerStyle)

	// === ROW 4+: DATA ===
	for i, conv := range conversions {
		row := i + 4
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), conv.ID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), conv.ProductId)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), conv.InitId)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), conv.FinalId)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), conv.ValueConv)
	}

	f.SetColWidth(sheet, "A", "A", 18)
	f.SetColWidth(sheet, "B", "B", 18)
	f.SetColWidth(sheet, "C", "C", 18)
	f.SetColWidth(sheet, "D", "D", 18)
	f.SetColWidth(sheet, "E", "E", 18)

	// Buat Table
	tableErr := f.AddTable(sheet, &excelize.Table{
		Range:             fmt.Sprintf("A3:E%d", len(conversions)+3),
		Name:              "UnitConversionsTable",
		StyleName:         "TableStyleMedium9",
		ShowFirstColumn:   false,
		ShowLastColumn:    false,
		ShowColumnStripes: false,
	})
	if tableErr != nil {
		log.Printf("[ExportUnitConversionsToExcel] AddTable warning: %v", tableErr)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write excel: %w", err)
	}

	return buf.Bytes(), nil
}
