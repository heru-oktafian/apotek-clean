package services

import (
	"fmt"
	"log"

	"github.com/heru-oktafian/fiber-apotek/models"
	"github.com/xuri/excelize/v2"
)

func (s *ExportServices) ExportFirstStockItemsToExcel(branchID string, firstStockID string) ([]byte, error) {
	var items []models.AllFirstStockItems

	// Query data first stock items dengan join ke products dan units
	query := s.db.Table("first_stock_items").
		Select("first_stock_items.id, first_stock_items.first_stock_id, first_stock_items.product_id, products.name as product_name, first_stock_items.price, first_stock_items.qty, units.name as unit_name, first_stock_items.sub_total").
		Joins("JOIN products ON products.id = first_stock_items.product_id").
		Joins("JOIN units ON units.id = products.unit_id").
		Where("first_stock_items.first_stock_id = ?", firstStockID)

	err := query.Order("products.name ASC").Find(&items).Error
	if err != nil {
		log.Printf("[ExportFirstStockItemsToExcel] Query error: %v", err)
		return nil, fmt.Errorf("failed to fetch first stock items: %w", err)
	}

	// Ambil header info
	var firstStock models.FirstStocks
	if err := s.db.Where("id = ? AND branch_id = ?", firstStockID, branchID).First(&firstStock).Error; err != nil {
		log.Printf("[ExportFirstStockItemsToExcel] FirstStock not found or mismatch branch: %v", err)
		return nil, fmt.Errorf("first stock not found or access denied")
	}

	f := excelize.NewFile()
	sheet := "Detail Stok Awal"
	f.SetSheetName("Sheet1", sheet)

	// === ROW 1: JUDUL ===
	f.SetCellValue(sheet, "A1", "LAPORAN DETAIL STOK AWAL")
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	f.MergeCell(sheet, "A1", "D1")
	f.SetCellStyle(sheet, "A1", "D1", titleStyle)
	f.SetRowHeight(sheet, 1, 30)

	// === ROW 2-5: INFO HEADER ===
	infoStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	f.SetCellValue(sheet, "A2", "ID STOK AWAL")
	f.SetCellValue(sheet, "B2", ": "+firstStock.ID)
	f.SetCellValue(sheet, "A3", "TANGGAL")
	f.SetCellValue(sheet, "B3", ": "+firstStock.FirstStockDate.Format("02/01/2006"))
	f.SetCellValue(sheet, "A4", "METODE PEMBAYARAN")
	f.SetCellValue(sheet, "B4", ": "+string(firstStock.Payment))
	f.SetCellValue(sheet, "A5", "DESKRIPSI")
	f.SetCellValue(sheet, "B5", ": "+firstStock.Description)
	f.SetCellStyle(sheet, "A2", "A5", infoStyle)

	// === ROW 7: HEADER TABEL ===
	headers := []string{"PRODUK", "QTY", "HARGA", "SUB TOTAL"}
	headerRow := 7
	for i, h := range headers {
		cell, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet, cell+fmt.Sprint(headerRow), h)
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	f.SetCellStyle(sheet, "A7", "D7", headerStyle)
	f.SetRowHeight(sheet, 7, 20)

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

	// === ROW 8+: DATA ===
	for i, item := range items {
		row := i + 8
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), item.ProductName)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("%d %s", item.Qty, item.UnitName))
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), formatRupiah(item.Price))
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), formatRupiah(item.SubTotal))

		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleLeft)
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styleRight)
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), styleRight)
	}

	// === BARIS TOTAL ===
	totalRow := len(items) + 8
	f.SetCellValue(sheet, fmt.Sprintf("A%d", totalRow), "TOTAL")
	f.MergeCell(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("C%d", totalRow))
	f.SetCellValue(sheet, fmt.Sprintf("D%d", totalRow), formatRupiah(firstStock.TotalFirstStock))

	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("D%d", totalRow), totalStyle)
	f.SetRowHeight(sheet, totalRow, 22)

	// === TABLE STYLE ===
	tableErr := f.AddTable(sheet, &excelize.Table{
		Range:             fmt.Sprintf("A7:D%d", len(items)+7),
		Name:              "FirstStockItemsTable",
		StyleName:         "TableStyleMedium9",
		ShowFirstColumn:   false,
		ShowLastColumn:    false,
		ShowColumnStripes: false,
	})
	if tableErr != nil {
		log.Printf("[ExportFirstStockItemsToExcel] AddTable warning: %v", tableErr)
	}

	// === LEBAR KOLOM ===
	f.SetColWidth(sheet, "A", "A", 40)
	f.SetColWidth(sheet, "B", "B", 15)
	f.SetColWidth(sheet, "C", "C", 20)
	f.SetColWidth(sheet, "D", "D", 20)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write excel: %w", err)
	}

	return buf.Bytes(), nil
}
