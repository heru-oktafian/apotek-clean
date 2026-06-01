package services

import (
	"fmt"
	"log"
	"strconv"

	"github.com/heru-oktafian/fiber-apotek/models"
	"github.com/xuri/excelize/v2"
)

// ExportProductsToExcel menghasilkan file Excel dengan data produk per branch
func (s *ExportServices) ExportProductsToExcel(branchID string) ([]byte, error) {
	var products []models.ProductDetail

	// Query dengan join agar dapat UnitName dan ProductCategoryName
	err := s.db.
		Select(`
			products.id,
			products.sku,
			products.name,
			products.alias,
			products.unit_id,
			units.name as unit_name,
			products.stock,
			products.purchase_price,
			products.sales_price,
			products.alternate_price,
			products.expired_date,
			products.product_category_id,
			product_categories.name as product_category_name
		`).
		Table("products").
		Joins("LEFT JOIN units ON units.id = products.unit_id").
		Joins("LEFT JOIN product_categories ON product_categories.id = products.product_category_id").
		Where("products.branch_id = ?", branchID).
		Order("products.name ASC").
		Scan(&products).Error

	if err != nil {
		log.Printf("[ExportProductsToExcel] Query error: %v", err)
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}

	f := excelize.NewFile()
	sheet := "Produk"
	f.SetSheetName("Sheet1", sheet)

	// === ROW 1: JUDUL ===
	f.SetCellValue(sheet, "A1", "DATA PRODUK")
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1565C0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	f.SetCellStyle(sheet, "A1", "J1", titleStyle)
	f.SetRowHeight(sheet, 1, 25)

	// === ROW 2: JARAK (kosong) ===
	// (tidak perlu action, biarkan kosong)

	// === ROW 3: HEADER ===
	headers := []string{
		"ID", "SKU", "NAME", "ALIAS",
		"PURCHASE PRI", "SALE PRI", "ALTERNATIF PRI",
		"STOCK", "UNIT", "EXPIRED DATE",
	}

	for i, h := range headers {
		cell, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet, cell+"3", h)
	}

	// Style Header (bold + background)
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
	f.SetCellStyle(sheet, "A3", "J3", headerStyle)

	// === ROW 4+: DATA ===
	for i, p := range products {
		row := i + 4

		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), p.ID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), p.SKU)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), p.Name)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), p.Alias)

		// Harga dengan format Rp. x.xxx
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), formatRupiah(p.PurchasePrice))
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), formatRupiah(p.SalesPrice))
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), formatRupiah(p.AlternatePrice))

		// STOCK + UNIT digabung seperti di screenshot kamu (500 PCS)
		stockUnit := fmt.Sprintf("%d %s", p.Stock, p.UnitName)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), stockUnit)

		// UNIT (nama satuan saja)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", row), p.UnitName)

		// Expired Date → DD/MM/YYYY
		f.SetCellValue(sheet, fmt.Sprintf("J%d", row), p.ExpiredDate.Format("02/01/2006"))
	}

	// Auto width kolom
	_ = f.SetColWidth(sheet, "A", "A", 18)
	_ = f.SetColWidth(sheet, "B", "B", 15)
	_ = f.SetColWidth(sheet, "C", "C", 35)
	_ = f.SetColWidth(sheet, "D", "D", 25)
	_ = f.SetColWidth(sheet, "E", "G", 15)
	_ = f.SetColWidth(sheet, "H", "H", 12)
	_ = f.SetColWidth(sheet, "I", "I", 10)
	_ = f.SetColWidth(sheet, "J", "J", 15)

	// Buat Table (supaya ada filter & styling bagus)
	tableErr := f.AddTable(sheet, &excelize.Table{
		Range:             fmt.Sprintf("A3:J%d", len(products)+3),
		Name:              "ProdukTable",
		StyleName:         "TableStyleMedium9",
		ShowFirstColumn:   false,
		ShowLastColumn:    false,
		ShowColumnStripes: false,
	})
	if tableErr != nil {
		log.Printf("[ExportProductsToExcel] AddTable warning (non-critical): %v", tableErr)
	}

	// Return sebagai byte (siap dikirim ke browser)
	buf, err := f.WriteToBuffer()
	if err != nil {
		log.Printf("[ExportProductsToExcel] WriteToBuffer error: %v", err)
		return nil, fmt.Errorf("failed to write excel: %w", err)
	}

	return buf.Bytes(), nil
}

// formatRupiah mengubah 4900 → "Rp. 4.900"
func formatRupiah(amount int) string {
	if amount == 0 {
		return "Rp. 0"
	}
	s := strconv.Itoa(amount)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "." + s[i:]
	}
	return "Rp. " + s
}
