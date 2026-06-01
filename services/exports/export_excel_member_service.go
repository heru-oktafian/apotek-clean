package services

import (
	"fmt"
	"log"

	"github.com/heru-oktafian/fiber-apotek/models"
	"github.com/xuri/excelize/v2"
)

func (s *ExportServices) ExportMembersToExcel(branchID string) ([]byte, error) {
	var members []models.Member

	err := s.db.Where("branch_id = ?", branchID).Order("name ASC").Find(&members).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch members: %w", err)
	}

	f := excelize.NewFile()
	sheet := "Members"
	f.SetSheetName("Sheet1", sheet)

	// === ROW 1: JUDUL ===
	f.SetCellValue(sheet, "A1", "DATA MEMBERS")
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1565C0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	f.SetCellStyle(sheet, "A1", "F1", titleStyle)
	f.SetRowHeight(sheet, 1, 25)

	// === ROW 2: JARAK (kosong) ===
	// (tidak perlu action, biarkan kosong)

	// === ROW 3: HEADER ===
	headers := []string{"MEMBER ID", "NAME", "CATEGORY ID", "PHONE", "ADDRESS", "POINTS"}
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
	f.SetCellStyle(sheet, "A3", "F3", headerStyle)

	// === ROW 4+: DATA ===
	for i, member := range members {
		row := i + 4
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("%d", member.ID))
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), member.Name)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), member.MemberCategoryId)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), member.Phone)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), member.Address)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), member.Points)
	}

	f.SetColWidth(sheet, "A", "A", 15)
	f.SetColWidth(sheet, "B", "B", 20)
	f.SetColWidth(sheet, "C", "C", 15)
	f.SetColWidth(sheet, "D", "D", 15)
	f.SetColWidth(sheet, "E", "E", 20)
	f.SetColWidth(sheet, "F", "F", 15)

	// Buat Table
	tableErr := f.AddTable(sheet, &excelize.Table{
		Range:             fmt.Sprintf("A3:F%d", len(members)+3),
		Name:              "MembersTable",
		StyleName:         "TableStyleMedium9",
		ShowFirstColumn:   false,
		ShowLastColumn:    false,
		ShowColumnStripes: false,
	})
	if tableErr != nil {
		log.Printf("[ExportMembersToExcel] AddTable warning: %v", tableErr)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write excel: %w", err)
	}

	return buf.Bytes(), nil
}
