package services

import (
	"fmt"
	"log"
	"time"

	"github.com/heru-oktafian/fiber-apotek/models"
	"github.com/xuri/excelize/v2"
)

func (s *ExportServices) ExportPurchasesToExcel(branchID string, month string) ([]byte, error) {
	var purchases []models.AllPurchases

	query := s.db.Table("purchases").
		Select("purchases.id, purchases.supplier_id, suppliers.name as supplier_name, purchases.purchase_date, purchases.total_purchase, purchases.payment").
		Joins("LEFT JOIN suppliers ON suppliers.id = purchases.supplier_id").
		Where("purchases.branch_id = ?", branchID)

	if month != "" {
		parsedTime, err := time.Parse("2006-01", month)
		if err == nil {
			startDate := parsedTime
			endDate := parsedTime.AddDate(0, 1, 0)
			query = query.Where("purchases.purchase_date >= ? AND purchases.purchase_date < ?", startDate, endDate)
		}
	}

	err := query.Order("purchases.purchase_date DESC").Find(&purchases).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch purchases: %w", err)
	}

	f := excelize.NewFile()
	sheet := "Purchases"
	f.SetSheetName("Sheet1", sheet)

	// === ROW 1: JUDUL ===
	f.SetCellValue(sheet, "A1", "PEMBELIAN "+month)
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	f.SetCellStyle(sheet, "A1", "E1", titleStyle)
	f.SetRowHeight(sheet, 1, 25)

	// === ROW 3: HEADER ===
	headers := []string{"ID", "SUPPLIER", "TANGGAL", "PEMBAYARAN", "TOTAL"}
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

	// === ROW 4+: DATA ===
	var grandTotal int
	for i, p := range purchases {
		row := i + 4
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), p.ID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), p.SupplierName)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), p.PurchaseDate.Format("02/01/2006"))
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), string(p.Payment))
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), formatRupiah(p.TotalPurchase))
		grandTotal += p.TotalPurchase

		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styleLeft)
		f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), styleRight)
	}

	// === BARIS GRAND TOTAL ===
	totalRow := len(purchases) + 4
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
		Range:             fmt.Sprintf("A3:E%d", len(purchases)+3),
		Name:              "PurchasesTable",
		StyleName:         "TableStyleMedium9",
		ShowFirstColumn:   false,
		ShowLastColumn:    false,
		ShowColumnStripes: false,
	})
	if tableErr != nil {
		log.Printf("[ExportPurchasesToExcel] AddTable warning: %v", tableErr)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write excel: %w", err)
	}

	return buf.Bytes(), nil
}
