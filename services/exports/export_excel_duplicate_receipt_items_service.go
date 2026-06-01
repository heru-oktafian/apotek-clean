package services

import (
	"fmt"
	"log"

	"github.com/heru-oktafian/fiber-apotek/models"
	"github.com/xuri/excelize/v2"
)

func (s *ExportServices) ExportDuplicateReceiptItemsToExcel(branchID string, duplicateReceiptID string) ([]byte, error) {
	var items []models.AllDuplicateReceiptItems

	// Query data duplicate receipt items dengan join ke products dan units
	query := s.db.Table("duplicate_receipt_items").
		Select("duplicate_receipt_items.id, duplicate_receipt_items.duplicate_receipt_id, duplicate_receipt_items.product_id, products.name as product_name, duplicate_receipt_items.price, duplicate_receipt_items.qty, units.name as unit_name, duplicate_receipt_items.sub_total").
		Joins("JOIN products ON products.id = duplicate_receipt_items.product_id").
		Joins("JOIN units ON units.id = products.unit_id").
		Where("duplicate_receipt_items.duplicate_receipt_id = ?", duplicateReceiptID)

	err := query.Order("products.name ASC").Find(&items).Error
	if err != nil {
		log.Printf("[ExportDuplicateReceiptItemsToExcel] Query error: %v", err)
		return nil, fmt.Errorf("failed to fetch duplicate receipt items: %w", err)
	}

	// Ambil header info
	var duplicateReceipt models.DuplicateReceipts
	if err := s.db.Where("id = ? AND branch_id = ?", duplicateReceiptID, branchID).First(&duplicateReceipt).Error; err != nil {
		log.Printf("[ExportDuplicateReceiptItemsToExcel] Duplicate Receipt not found or mismatch branch: %v", err)
		return nil, fmt.Errorf("duplicate receipt not found or access denied")
	}

	f := excelize.NewFile()
	sheet := "Detail Kopi Resep"
	f.SetSheetName("Sheet1", sheet)

	// === ROW 1: JUDUL ===
	f.SetCellValue(sheet, "A1", "LAPORAN DETAIL KOPI RESEP")
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	f.MergeCell(sheet, "A1", "D1")
	f.SetCellStyle(sheet, "A1", "D1", titleStyle)
	f.SetRowHeight(sheet, 1, 30)

	// === ROW 2-4: INFO HEADER ===
	infoStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	f.SetCellValue(sheet, "A2", "ID KOPI RESEP")
	f.SetCellValue(sheet, "B2", ": "+duplicateReceipt.ID)
	f.SetCellValue(sheet, "A3", "TANGGAL")
	f.SetCellValue(sheet, "B3", ": "+duplicateReceipt.DuplicateReceiptDate.Format("02/01/2006"))
	f.SetCellValue(sheet, "A4", "METODE PEMBAYARAN")
	f.SetCellValue(sheet, "B4", ": "+string(duplicateReceipt.Payment))
	f.SetCellStyle(sheet, "A2", "A4", infoStyle)

	// === ROW 6: HEADER TABEL ===
	headers := []string{"PRODUK", "HARGA", "JUMLAH", "SUB TOTAL"}
	headerRow := 6
	for i, h := range headers {
		cell, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet, cell+fmt.Sprint(headerRow), h)
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	f.SetCellStyle(sheet, "A6", "D6", headerStyle)
	f.SetRowHeight(sheet, 6, 20)

	// === STYLE DATA ===
	styleLeft, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	styleCenter, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	styleRight, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})

	// === ROW 7+: DATA ===
	for i, item := range items {
		row := i + 7
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), item.ProductName)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), formatRupiah(item.Price))
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("%d %s", item.Qty, item.UnitName))
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), formatRupiah(item.SubTotal))

		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleLeft)
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styleRight)
		f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), styleRight)
	}

	// === BARIS TOTAL ===
	totalRow := len(items) + 7
	f.SetCellValue(sheet, fmt.Sprintf("A%d", totalRow), "TOTAL")
	f.MergeCell(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("C%d", totalRow))
	f.SetCellValue(sheet, fmt.Sprintf("D%d", totalRow), formatRupiah(duplicateReceipt.TotalDuplicateReceipt))

	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("D%d", totalRow), totalStyle)
	f.SetRowHeight(sheet, totalRow, 22)

	// === TABLE STYLE ===
	tableErr := f.AddTable(sheet, &excelize.Table{
		Range:             fmt.Sprintf("A6:D%d", len(items)+6),
		Name:              "DuplicateReceiptItemsTable",
		StyleName:         "TableStyleMedium9",
		ShowFirstColumn:   false,
		ShowLastColumn:    false,
		ShowColumnStripes: false,
	})
	if tableErr != nil {
		log.Printf("[ExportDuplicateReceiptItemsToExcel] AddTable warning: %v", tableErr)
	}

	// === LEBAR KOLOM ===
	f.SetColWidth(sheet, "A", "A", 40)
	f.SetColWidth(sheet, "B", "B", 20)
	f.SetColWidth(sheet, "C", "C", 15)
	f.SetColWidth(sheet, "D", "D", 20)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write excel: %w", err)
	}

	return buf.Bytes(), nil
}
