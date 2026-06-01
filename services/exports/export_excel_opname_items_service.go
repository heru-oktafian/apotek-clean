package services

import (
	"fmt"
	"log"

	"github.com/heru-oktafian/fiber-apotek/models"
	"github.com/xuri/excelize/v2"
)

// OpnameItemExcel adalah struct lokal untuk menampung hasil query opname items
type OpnameItemExcel struct {
	ProductName string
	UnitName    string
	QtySystem   int
	QtyPhysical int
	QtyDiff     int
	Price       int
	SubTotal    int
}

func (s *ExportServices) ExportOpnameItemsToExcel(branchID string, opnameID string) ([]byte, error) {
	var items []OpnameItemExcel

	// Query data opname items dengan join ke products dan units
	// Menghitung selisih (diff) antara Qty (Fisik) dan QtyExist (Sistem)
	query := s.db.Table("opname_items").
		Select("products.name as product_name, units.name as unit_name, opname_items.qty_exist as qty_system, opname_items.qty as qty_physical, (opname_items.qty - opname_items.qty_exist) as qty_diff, opname_items.price, opname_items.sub_total").
		Joins("JOIN products ON products.id = opname_items.product_id").
		Joins("JOIN units ON units.id = products.unit_id").
		Where("opname_items.opname_id = ?", opnameID)

	err := query.Order("products.name ASC").Scan(&items).Error
	if err != nil {
		log.Printf("[ExportOpnameItemsToExcel] Query error: %v", err)
		return nil, fmt.Errorf("failed to fetch opname items: %w", err)
	}

	// Ambil header info
	var opname models.Opnames
	if err := s.db.Where("id = ? AND branch_id = ?", opnameID, branchID).First(&opname).Error; err != nil {
		log.Printf("[ExportOpnameItemsToExcel] Opname not found or mismatch branch: %v", err)
		return nil, fmt.Errorf("opname not found or access denied")
	}

	f := excelize.NewFile()
	sheet := "Detail Opname"
	f.SetSheetName("Sheet1", sheet)

	// === ROW 1: JUDUL ===
	f.SetCellValue(sheet, "A1", "LAPORAN DETAIL OPNAME")
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	f.MergeCell(sheet, "A1", "F1")
	f.SetCellStyle(sheet, "A1", "F1", titleStyle)
	f.SetRowHeight(sheet, 1, 30)

	// === ROW 2-5: INFO HEADER ===
	infoStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	f.SetCellValue(sheet, "A2", "ID OPNAME")
	f.SetCellValue(sheet, "B2", ": "+opname.ID)
	f.SetCellValue(sheet, "A3", "TANGGAL")
	f.SetCellValue(sheet, "B3", ": "+opname.OpnameDate.Format("02/01/2006"))
	f.SetCellValue(sheet, "A4", "METODE PEMBAYARAN")
	f.SetCellValue(sheet, "B4", ": "+string(opname.Payment))
	f.SetCellValue(sheet, "A5", "KETERANGAN")
	f.SetCellValue(sheet, "B5", ": "+opname.Description)
	f.SetCellStyle(sheet, "A2", "A5", infoStyle)

	// === ROW 7: HEADER TABEL ===
	headers := []string{"PRODUK", "SISTEM", "FISIK", "SELISIH", "HARGA", "SUB TOTAL"}
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
	f.SetCellStyle(sheet, "A7", "F7", headerStyle)
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
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("%d %s", item.QtySystem, item.UnitName))
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("%d %s", item.QtyPhysical, item.UnitName))
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("%d %s", item.QtyDiff, item.UnitName))
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), formatRupiah(item.Price))
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), formatRupiah(item.SubTotal))

		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleLeft)
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), styleRight)
		f.SetCellStyle(sheet, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), styleRight)
	}

	// === BARIS TOTAL SELISIH ===
	totalRow := len(items) + 8
	f.SetCellValue(sheet, fmt.Sprintf("A%d", totalRow), "TOTAL SELISIH")
	f.MergeCell(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("E%d", totalRow))
	f.SetCellValue(sheet, fmt.Sprintf("F%d", totalRow), formatRupiah(opname.TotalOpname))

	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("F%d", totalRow), totalStyle)
	f.SetRowHeight(sheet, totalRow, 22)

	// === TABLE STYLE ===
	tableErr := f.AddTable(sheet, &excelize.Table{
		Range:             fmt.Sprintf("A7:F%d", len(items)+7),
		Name:              "OpnameItemsTable",
		StyleName:         "TableStyleMedium9",
		ShowFirstColumn:   false,
		ShowLastColumn:    false,
		ShowColumnStripes: false,
	})
	if tableErr != nil {
		log.Printf("[ExportOpnameItemsToExcel] AddTable warning: %v", tableErr)
	}

	// === LEBAR KOLOM ===
	f.SetColWidth(sheet, "A", "A", 35)
	f.SetColWidth(sheet, "B", "B", 12)
	f.SetColWidth(sheet, "C", "C", 12)
	f.SetColWidth(sheet, "D", "D", 12)
	f.SetColWidth(sheet, "E", "E", 20)
	f.SetColWidth(sheet, "F", "F", 20)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write excel: %w", err)
	}

	return buf.Bytes(), nil
}
