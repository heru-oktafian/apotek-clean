package services

import (
	"fmt"
	"log"

	"github.com/heru-oktafian/fiber-apotek/models"
	"github.com/xuri/excelize/v2"
)

func (s *ExportServices) ExportSaleReturnItemsToExcel(branchID string, saleReturnID string) ([]byte, error) {
	var items []models.AllSaleReturnItems

	// Query data sale return items dengan join ke products dan units
	query := s.db.Table("sale_return_items").
		Select("sale_return_items.id, sale_return_items.sale_return_id, sale_return_items.product_id as pro_id, products.name as pro_name, sale_return_items.price, sale_return_items.qty, units.id as unit_id, units.name as unit_name, sale_return_items.sub_total, sale_return_items.expired_date").
		Joins("JOIN products ON products.id = sale_return_items.product_id").
		Joins("JOIN units ON units.id = products.unit_id").
		Where("sale_return_items.sale_return_id = ?", saleReturnID)

	err := query.Order("products.name ASC").Find(&items).Error
	if err != nil {
		log.Printf("[ExportSaleReturnItemsToExcel] Query error: %v", err)
		return nil, fmt.Errorf("failed to fetch sale return items: %w", err)
	}

	// Ambil header info
	var saleReturn models.SaleReturns
	if err := s.db.Where("id = ? AND branch_id = ?", saleReturnID, branchID).First(&saleReturn).Error; err != nil {
		log.Printf("[ExportSaleReturnItemsToExcel] SaleReturn not found or mismatch branch: %v", err)
		return nil, fmt.Errorf("sale return not found or access denied")
	}

	f := excelize.NewFile()
	sheet := "Detail Retur Penjualan"
	f.SetSheetName("Sheet1", sheet)

	// === ROW 1: JUDUL ===
	f.SetCellValue(sheet, "A1", "LAPORAN DETAIL RETUR PENJUALAN")
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	f.MergeCell(sheet, "A1", "E1")
	f.SetCellStyle(sheet, "A1", "E1", titleStyle)
	f.SetRowHeight(sheet, 1, 30)

	// === ROW 2-5: INFO HEADER ===
	infoStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	f.SetCellValue(sheet, "A2", "ID RETUR PENJUALAN")
	f.SetCellValue(sheet, "B2", ": "+saleReturn.ID)
	f.SetCellValue(sheet, "A3", "TANGGAL RETUR")
	f.SetCellValue(sheet, "B3", ": "+saleReturn.ReturnDate.Format("02/01/2006"))
	f.SetCellValue(sheet, "A4", "ID PENJUALAN")
	f.SetCellValue(sheet, "B4", ": "+saleReturn.SaleId)
	f.SetCellValue(sheet, "A5", "METODE PEMBAYARAN")
	f.SetCellValue(sheet, "B5", ": "+string(saleReturn.Payment))
	f.SetCellStyle(sheet, "A2", "A5", infoStyle)

	// === ROW 7: HEADER TABEL ===
	headers := []string{"PRODUK", "KADALUARSA", "HARGA", "QTY", "SUB TOTAL"}
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
	f.SetCellStyle(sheet, "A7", "E7", headerStyle)
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
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), item.ProName)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), item.ExpiredDate.Format("02/01/2006"))
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), formatRupiah(item.Price))
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("%d %s", item.Qty, item.UnitName))
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), formatRupiah(item.SubTotal))

		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleLeft)
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styleRight)
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), styleRight)
	}

	// === BARIS TOTAL ===
	totalRow := len(items) + 8
	f.SetCellValue(sheet, fmt.Sprintf("A%d", totalRow), "TOTAL")
	f.MergeCell(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("D%d", totalRow))
	f.SetCellValue(sheet, fmt.Sprintf("E%d", totalRow), formatRupiah(saleReturn.TotalReturn))

	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("E%d", totalRow), totalStyle)
	f.SetRowHeight(sheet, totalRow, 22)

	// === TABLE STYLE ===
	tableErr := f.AddTable(sheet, &excelize.Table{
		Range:             fmt.Sprintf("A7:E%d", len(items)+7),
		Name:              "SaleReturnItemsTable",
		StyleName:         "TableStyleMedium9",
		ShowFirstColumn:   false,
		ShowLastColumn:    false,
		ShowColumnStripes: false,
	})
	if tableErr != nil {
		log.Printf("[ExportSaleReturnItemsToExcel] AddTable warning: %v", tableErr)
	}

	// === LEBAR KOLOM ===
	f.SetColWidth(sheet, "A", "A", 35)
	f.SetColWidth(sheet, "B", "B", 15)
	f.SetColWidth(sheet, "C", "C", 20)
	f.SetColWidth(sheet, "D", "D", 15)
	f.SetColWidth(sheet, "E", "E", 20)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write excel: %w", err)
	}

	return buf.Bytes(), nil
}
