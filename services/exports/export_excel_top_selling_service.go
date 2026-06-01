package services

import (
	"fmt"
	"log"
	"time"

	"github.com/heru-oktafian/fiber-apotek/configs"
	"github.com/xuri/excelize/v2"
)

// ExportTopSellingToExcel membuat Excel untuk laporan produk terlaris (1 bulan terakhir)
func (s *ExportServices) ExportTopSellingToExcel(branchID string) ([]byte, error) {
	now := time.Now().In(configs.Location)
	oneMonthAgo := now.AddDate(0, -1, 0)

	type Result struct {
		ProductID string `gorm:"column:product_id"`
		Name      string `gorm:"column:name"`
		TotalQty  int    `gorm:"column:total_qty"`
	}

	var results []Result
	err := s.db.
		Table("sale_items").
		Select("products.id as product_id, products.name, SUM(sale_items.qty) as total_qty").
		Joins("JOIN sales ON sales.id = sale_items.sale_id").
		Joins("JOIN products ON products.id = sale_items.product_id").
		Where("sales.sale_date >= ? AND sales.branch_id = ?", oneMonthAgo, branchID).
		Group("products.id, products.name").
		Order("total_qty DESC").
		Limit(50).
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch top selling products: %w", err)
	}

	f := excelize.NewFile()
	sheet := "Fast Moving"
	f.SetSheetName("Sheet1", sheet)

	f.SetCellValue(sheet, "A1", fmt.Sprintf("PRODUK FAST MOVING - %s", now.Format("2006-01-02")))
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	f.SetCellStyle(sheet, "A1", "D1", titleStyle)
	f.MergeCell(sheet, "A1", "D1")
	f.SetRowHeight(sheet, 1, 25)

	headers := []string{"No", "PRODUCT ID", "NAME", "TOTAL QTY"}
	for i, h := range headers {
		cell, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet, cell+"3", h)
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	f.SetCellStyle(sheet, "A3", "D3", headerStyle)

	styleCenter, _ := f.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"}})
	styleLeft, _ := f.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"}})
	styleRight, _ := f.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"}})

	var totalQty int
	for i, r := range results {
		row := i + 4
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), i+1)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), r.ProductID)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), r.Name)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), r.TotalQty)

		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styleLeft)
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), styleRight)

		totalQty += r.TotalQty
	}

	totalRow := len(results) + 4
	f.SetCellValue(sheet, fmt.Sprintf("A%d", totalRow), "TOTAL")
	f.MergeCell(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("C%d", totalRow))
	f.SetCellValue(sheet, fmt.Sprintf("D%d", totalRow), totalQty)

	grandStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("D%d", totalRow), grandStyle)

	f.SetColWidth(sheet, "A", "A", 8)
	f.SetColWidth(sheet, "B", "B", 18)
	f.SetColWidth(sheet, "C", "C", 40)
	f.SetColWidth(sheet, "D", "D", 15)

	if len(results) > 0 {
		if err := f.AddTable(sheet, &excelize.Table{Range: fmt.Sprintf("A3:D%d", len(results)+3), Name: "TopSellingTable", StyleName: "TableStyleMedium9"}); err != nil {
			log.Printf("[ExportTopSellingToExcel] AddTable warning: %v", err)
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write excel: %w", err)
	}
	return buf.Bytes(), nil
}
