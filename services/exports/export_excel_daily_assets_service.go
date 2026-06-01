package services

import (
	"fmt"
	"log"
	"time"

	"github.com/heru-oktafian/fiber-apotek/models"
	"github.com/xuri/excelize/v2"
)

func (s *ExportServices) ExportDailyAssetsToExcel(branchID string, month string) ([]byte, error) {
	var assets []models.DailyAsset

	query := s.db.Where("branch_id = ?", branchID)

	if month != "" {
		parsedTime, err := time.Parse("2006-01", month)
		if err == nil {
			startDate := parsedTime
			endDate := parsedTime.AddDate(0, 1, 0)
			query = query.Where("asset_date >= ? AND asset_date < ?", startDate, endDate)
		}
	}

	err := query.Order("asset_date DESC").Find(&assets).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch daily assets: %w", err)
	}

	f := excelize.NewFile()
	sheet := "Daily Assets"
	f.SetSheetName("Sheet1", sheet)

	// === ROW 1: JUDUL ===
	f.SetCellValue(sheet, "A1", "LAPORAN ASET "+month)
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	f.SetCellStyle(sheet, "A1", "C1", titleStyle)
	f.MergeCell(sheet, "A1", "C1")
	f.SetRowHeight(sheet, 1, 25)

	// === ROW 3: HEADER ===
	headers := []string{"ID", "TANGGAL", "NILAI ASET"}
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
	f.SetCellStyle(sheet, "A3", "C3", headerStyle)

	styleCenter, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	styleRight, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})

	// === ROW 4+: DATA ===
	// var totalAssetValue int
	for i, a := range assets {
		row := i + 4
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), a.ID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), a.AssetDate.Format("02/01/2006"))
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), formatRupiah(a.AssetValue))
		// totalAssetValue += a.AssetValue

		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styleRight)
	}

	// === BARIS ASET TERBARU ===
	totalRow := len(assets) + 4
	f.SetCellValue(sheet, fmt.Sprintf("A%d", totalRow), "ASET TERBARU")
	f.MergeCell(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("B%d", totalRow))

	var latestAsset int
	if len(assets) > 0 {
		// Karena diurutkan DESC pada line 26, index 0 adalah data terbaru
		latestAsset = assets[0].AssetValue
	}
	f.SetCellValue(sheet, fmt.Sprintf("C%d", totalRow), formatRupiah(latestAsset))
	// No total for average usually, or we can just leave it blank

	grandTotalStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF", Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("C%d", totalRow), grandTotalStyle)
	f.SetRowHeight(sheet, totalRow, 20)

	f.SetColWidth(sheet, "A", "A", 18)
	f.SetColWidth(sheet, "B", "B", 15)
	f.SetColWidth(sheet, "C", "C", 20)

	tableErr := f.AddTable(sheet, &excelize.Table{
		Range:             fmt.Sprintf("A3:C%d", len(assets)+3),
		Name:              "DailyAssetsTable",
		StyleName:         "TableStyleMedium9",
		ShowFirstColumn:   false,
		ShowLastColumn:    false,
		ShowColumnStripes: false,
	})
	if tableErr != nil {
		log.Printf("[ExportDailyAssetsToExcel] AddTable warning: %v", tableErr)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write excel: %w", err)
	}

	return buf.Bytes(), nil
}
