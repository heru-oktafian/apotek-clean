package services

import (
	"fmt"
	"log"
	"time"

	"github.com/heru-oktafian/fiber-apotek/configs"
	"github.com/xuri/excelize/v2"
)

// ExportLeastSellingToExcel membuat Excel untuk laporan produk dengan penjualan terendah
func (s *ExportServices) ExportLeastSellingToExcel(branchID string) ([]byte, error) {
	now := time.Now().In(configs.Location)
	oneMonthAgo := now.AddDate(0, -1, 0)

	// Subquery: total sold per product
	subQuery := s.db.
		Table("sale_items").
		Select("product_id, SUM(qty) as total_sold").
		Joins("JOIN sales ON sales.id = sale_items.sale_id").
		Where("sales.sale_date BETWEEN ? AND ?", oneMonthAgo, now).
		Group("product_id")

	type Result struct {
		ProductID   string `gorm:"column:product_id"`
		ProductName string `gorm:"column:product_name"`
		Stock       int    `gorm:"column:stock"`
		TotalSold   int    `gorm:"column:total_sold"`
	}

	var results []Result
	if err := s.db.
		Table("products p").
		Select("p.id as product_id, p.name as product_name, p.stock, COALESCE(s.total_sold, 0) as total_sold").
		Joins("LEFT JOIN (?) as s ON p.id = s.product_id", subQuery).
		Where("p.stock >= ? AND p.branch_id = ?", 1, branchID).
		Order("total_sold ASC").
		Limit(50).
		Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch least selling products: %w", err)
	}

	f := excelize.NewFile()
	sheet := "Slow Moving"
	f.SetSheetName("Sheet1", sheet)

	f.SetCellValue(sheet, "A1", fmt.Sprintf("PRODUK SLOW MOVING - %s", now.Format("2006-01-02")))
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	f.SetCellStyle(sheet, "A1", "E1", titleStyle)
	f.MergeCell(sheet, "A1", "E1")
	f.SetRowHeight(sheet, 1, 25)

	headers := []string{"No", "PRODUCT ID", "NAME", "STOCK", "TOTAL SOLD"}
	for i, h := range headers {
		cell, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet, cell+"3", h)
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	f.SetCellStyle(sheet, "A3", "E3", headerStyle)

	styleCenter, _ := f.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"}})
	styleLeft, _ := f.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"}})
	styleRight, _ := f.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"}})

	var totalSoldSum int
	for i, r := range results {
		row := i + 4
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), i+1)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), r.ProductID)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), r.ProductName)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), r.Stock)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), r.TotalSold)

		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styleLeft)
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), styleRight)

		totalSoldSum += r.TotalSold
	}

	totalRow := len(results) + 4
	f.SetCellValue(sheet, fmt.Sprintf("A%d", totalRow), "TOTAL")
	f.MergeCell(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("D%d", totalRow))
	f.SetCellValue(sheet, fmt.Sprintf("E%d", totalRow), totalSoldSum)

	grandStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", totalRow), fmt.Sprintf("E%d", totalRow), grandStyle)

	f.SetColWidth(sheet, "A", "A", 8)
	f.SetColWidth(sheet, "B", "B", 18)
	f.SetColWidth(sheet, "C", "C", 40)
	f.SetColWidth(sheet, "D", "D", 12)
	f.SetColWidth(sheet, "E", "E", 15)

	if len(results) > 0 {
		if err := f.AddTable(sheet, &excelize.Table{Range: fmt.Sprintf("A3:E%d", len(results)+3), Name: "LeastSellingTable", StyleName: "TableStyleMedium9"}); err != nil {
			log.Printf("[ExportLeastSellingToExcel] AddTable warning: %v", err)
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write excel: %w", err)
	}
	return buf.Bytes(), nil
}
