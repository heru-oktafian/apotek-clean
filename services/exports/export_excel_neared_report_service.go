package services

import (
	"fmt"
	"log"
	"time"

	"github.com/heru-oktafian/fiber-apotek/configs"
	"github.com/xuri/excelize/v2"
)

// ExportNearedReportToExcel menghasilkan file Excel berisi produk yang mendekati
// masa kadaluarsa (<= 3 bulan). Desain mengikuti pola pada ExportDailyAssetsToExcel:
// - Judul di baris 1
// - Header di baris 3
// - Data mulai baris 4
// - Baris ringkasan akhir menampilkan jumlah produk yang ter-list
func (s *ExportServices) ExportNearedReportToExcel(branchID string) ([]byte, error) {
	// Gunakan timezone dari configs agar konsisten dengan controller
	nowWIB := time.Now().In(configs.Location)
	threeMonthsLater := nowWIB.AddDate(0, 3, 0)

	type ProductQueryResult struct {
		ID          string    `gorm:"column:id"`
		SKU         string    `gorm:"column:sku"`
		Name        string    `gorm:"column:name"`
		Stock       int       `gorm:"column:stock"`
		Unit        string    `gorm:"column:unit"`
		ExpiredDate time.Time `gorm:"column:expired_date"`
	}

	var rawProducts []ProductQueryResult

	err := s.db.Table("products").
		Select("products.id, products.sku, products.name, products.stock, units.name as unit, products.expired_date").
		Joins("LEFT JOIN units ON products.unit_id = units.id").
		Where("products.expired_date <= ? AND products.stock >= ? AND products.branch_id = ?", threeMonthsLater, 1, branchID).
		Order("products.expired_date ASC").
		Scan(&rawProducts).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch neared products: %w", err)
	}

	f := excelize.NewFile()
	sheet := "Near Expired"
	f.SetSheetName("Sheet1", sheet)

	// Title (sertakan tanggal pembuatan laporan)
	f.SetCellValue(sheet, "A1", fmt.Sprintf("LAPORAN PRODUK MENDEKATI KADALUARSA - %s", nowWIB.Format("2006-01-02")))
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	f.SetCellStyle(sheet, "A1", "F1", titleStyle)
	f.MergeCell(sheet, "A1", "F1")
	f.SetRowHeight(sheet, 1, 25)

	// Header
	headers := []string{"ID", "SKU", "NAMA PRODUK", "STOK", "UNIT", "TANGGAL KADALUARSA"}
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
	f.SetCellStyle(sheet, "A3", "F3", headerStyle)

	styleCenter, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	styleRight, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})

	// Data rows
	for i, p := range rawProducts {
		row := i + 4
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), p.ID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), p.SKU)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), p.Name)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), p.Stock)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), p.Unit)
		// Tampilkan tanggal kadaluarsa dalam format DD/MM/YYYY untuk tampilan lokal
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), p.ExpiredDate.Format("02/01/2006"))

		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styleRight)
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), styleCenter)
	}

	// Summary row: jumlah produk
	totalRow := len(rawProducts) + 4
	f.SetCellValue(sheet, fmt.Sprintf("A%d", totalRow), "TOTAL PRODUK")
	f.MergeCell(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("E%d", totalRow))
	f.SetCellValue(sheet, fmt.Sprintf("F%d", totalRow), fmt.Sprintf("%d", len(rawProducts)))

	grandTotalStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF", Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("F%d", totalRow), grandTotalStyle)
	f.SetRowHeight(sheet, totalRow, 20)

	f.SetColWidth(sheet, "A", "A", 18)
	f.SetColWidth(sheet, "B", "B", 18)
	f.SetColWidth(sheet, "C", "C", 40)
	f.SetColWidth(sheet, "D", "D", 12)
	f.SetColWidth(sheet, "E", "E", 15)
	f.SetColWidth(sheet, "F", "F", 18)

	// Tambahkan Excel table hanya jika ada minimal 1 baris data.
	if len(rawProducts) > 0 {
		tableErr := f.AddTable(sheet, &excelize.Table{
			Range:             fmt.Sprintf("A3:F%d", len(rawProducts)+3),
			Name:              "NearedReportTable",
			StyleName:         "TableStyleMedium9",
			ShowFirstColumn:   false,
			ShowLastColumn:    false,
			ShowColumnStripes: false,
		})
		if tableErr != nil {
			log.Printf("[ExportNearedReportToExcel] AddTable warning: %v", tableErr)
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write excel: %w", err)
	}

	return buf.Bytes(), nil
}
