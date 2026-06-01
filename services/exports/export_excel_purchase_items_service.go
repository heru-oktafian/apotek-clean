package services

import (
	"fmt"
	"log"

	"github.com/heru-oktafian/fiber-apotek/models"
	"github.com/xuri/excelize/v2"
)

func (s *ExportServices) ExportPurchaseItemsToExcel(branchID string, purchaseID string) ([]byte, error) {
	var items []models.AllPurchaseItems

	// Query data purchase items
	query := s.db.Table("purchase_items").
		Select("purchase_items.id, purchase_items.purchase_id, purchase_items.product_id, products.name as product_name, purchase_items.price, purchase_items.qty, purchase_items.unit_id, units.name as unit_name, purchase_items.sub_total, purchase_items.expired_date").
		Joins("JOIN products ON products.id = purchase_items.product_id").
		Joins("JOIN units ON units.id = purchase_items.unit_id").
		Where("purchase_items.purchase_id = ?", purchaseID)

	err := query.Order("products.name ASC").Find(&items).Error
	if err != nil {
		log.Printf("[ExportPurchaseItemsToExcel] Query error: %v", err)
		return nil, fmt.Errorf("failed to fetch purchase items: %w", err)
	}

	// Ambil header info
	var purchase models.AllPurchases
	err = s.db.Table("purchases").
		Select("purchases.id, purchases.supplier_id, suppliers.name as supplier_name, purchases.purchase_date, purchases.total_purchase, purchases.payment").
		Joins("JOIN suppliers ON suppliers.id = purchases.supplier_id").
		Where("purchases.id = ? AND purchases.branch_id = ?", purchaseID, branchID).
		First(&purchase).Error

	if err != nil {
		log.Printf("[ExportPurchaseItemsToExcel] Purchase not found or mismatch branch: %v", err)
		return nil, fmt.Errorf("purchase not found or access denied")
	}

	f := excelize.NewFile()
	sheet := "Detail Pembelian"
	f.SetSheetName("Sheet1", sheet)

	// === HEADER INFO ===
	// Title
	f.SetCellValue(sheet, "A1", "LAPORAN DETAIL PEMBELIAN")
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	f.MergeCell(sheet, "A1", "E1")
	f.SetCellStyle(sheet, "A1", "E1", titleStyle)
	f.SetRowHeight(sheet, 1, 30)

	// Info Rows
	f.SetCellValue(sheet, "A2", "ID PEMBELIAN")
	f.SetCellValue(sheet, "B2", ": "+purchase.ID)
	f.SetCellValue(sheet, "A3", "TANGGAL")
	f.SetCellValue(sheet, "B3", ": "+purchase.PurchaseDate.Format("02/01/2006"))
	f.SetCellValue(sheet, "A4", "SUPPLIER")
	f.SetCellValue(sheet, "B4", ": "+purchase.SupplierName)
	f.SetCellValue(sheet, "A5", "METODE PEMBAYARAN")
	f.SetCellValue(sheet, "B5", ": "+string(purchase.Payment))

	infoStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	f.SetCellStyle(sheet, "A2", "A5", infoStyle)

	// === TABLE HEADER ===
	headers := []string{"PRODUK", "EXPIRED", "JUMLAH", "HARGA", "SUB TOTAL"}
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

	// === DATA ROWS ===
	styleCenter, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	styleLeft, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	styleRight, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})

	for i, item := range items {
		row := i + 8
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), item.ProductName)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), item.ExpiredDate.Format("02/01/2006"))
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("%d %s", item.Qty, item.UnitName))
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), formatRupiah(item.Price))
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), formatRupiah(item.SubTotal))

		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleLeft)
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), styleRight)
		f.SetCellStyle(sheet, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), styleRight)
	}

	// === TOTAL ROW ===
	totalRow := len(items) + 8
	f.SetCellValue(sheet, fmt.Sprintf("A%d", totalRow), "TOTAL")
	f.MergeCell(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("D%d", totalRow))
	f.SetCellValue(sheet, fmt.Sprintf("E%d", totalRow), formatRupiah(purchase.TotalPurchase))

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
		Name:              "PurchaseItemsTable",
		StyleName:         "TableStyleMedium9",
		ShowFirstColumn:   false,
		ShowLastColumn:    false,
		ShowColumnStripes: false,
	})
	if tableErr != nil {
		log.Printf("[ExportPurchaseItemsToExcel] AddTable warning: %v", tableErr)
	}

	// Column Widths
	f.SetColWidth(sheet, "A", "A", 35)
	f.SetColWidth(sheet, "B", "B", 15)
	f.SetColWidth(sheet, "C", "C", 15)
	f.SetColWidth(sheet, "D", "D", 20)
	f.SetColWidth(sheet, "E", "E", 20)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write excel: %w", err)
	}

	return buf.Bytes(), nil
}
