package services

import (
	"fmt"
	"time"

	"github.com/heru-oktafian/fiber-apotek/models"
	"github.com/xuri/excelize/v2"
)

func (s *ExportServices) ExportNeracaSaldoToExcel(branchID string, month string) ([]byte, error) {
	// Konversi bulan ke rentang tanggal
	var startDate, endDate time.Time
	var err error
	if month != "" {
		startDate, err = time.Parse("2006-01", month)
		if err != nil {
			return nil, fmt.Errorf("invalid month format: %w", err)
		}
		endDate = startDate.AddDate(0, 1, 0)
	}

	type Summary struct {
		TransactionType string
		TransactionDate string
		Total           int
	}

	var summaries []Summary

	query := s.db.Table("transaction_reports").
		Select("transaction_type, DATE(created_at) AS transaction_date, SUM(total) AS total").
		Where("branch_id = ? AND payment != 'paid_by_credit'", branchID).
		Group("transaction_type, DATE(created_at)").
		Order("transaction_date ASC")

	if month != "" {
		query = query.Where("created_at >= ? AND created_at < ?", startDate, endDate)
	}

	if err := query.Scan(&summaries).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch transaction summaries: %w", err)
	}

	f := excelize.NewFile()
	sheet := "Neraca Saldo"
	f.SetSheetName("Sheet1", sheet)

	// === ROW 1: JUDUL ===
	f.SetCellValue(sheet, "A1", "NERACA SALDO "+month)
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	f.SetCellStyle(sheet, "A1", "E1", titleStyle)
	f.MergeCell(sheet, "A1", "E1")
	f.SetRowHeight(sheet, 1, 25)

	// Setup Styles
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
	styleCenter, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	styleLeft, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	styleRight, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	grandTotalStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF", Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1E88E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})

	// Separate Debit and Credit and build unified rows
	var rows []struct {
		TransactionType string
		TransactionDate string
		Debit           int
		Credit          int
	}
	var totalDebit, totalCredit int

	for _, s := range summaries {
		r := struct {
			TransactionType string
			TransactionDate string
			Debit           int
			Credit          int
		}{
			TransactionType: s.TransactionType,
			TransactionDate: s.TransactionDate,
		}
		switch s.TransactionType {
		case string(models.Sale), string(models.Income), string(models.BuyReturn):
			r.Debit = s.Total
			totalDebit += s.Total
		case string(models.Purchase), string(models.Expense), string(models.SaleReturn):
			r.Credit = s.Total
			totalCredit += s.Total
		}
		rows = append(rows, r)
	}

	currentRow := 3
	// --- TOP SUMMARY TEXT ---
	summaryText := fmt.Sprintf("Debit : %s   Kredit : %s   Sisa Estimasi Saldo : %s", formatRupiah(totalDebit), formatRupiah(totalCredit), formatRupiah(totalDebit-totalCredit))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", currentRow), summaryText)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("E%d", currentRow), styleLeft)
	f.MergeCell(sheet, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("E%d", currentRow))
	currentRow += 2

	// --- TABLE HEADER ---
	headers := []string{"No", "TANGGAL", "TRANSAKSI", "DEBET", "KREDIT"}
	for i, h := range headers {
		cell, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet, cell+fmt.Sprintf("%d", currentRow), h)
	}
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("E%d", currentRow), headerStyle)
	currentRow++

	for idx, r := range rows {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", currentRow), idx+1)
		parsedDate, _ := time.Parse("2006-01-02", r.TransactionDate)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", currentRow), parsedDate.Format("02/01/2006"))
		f.SetCellValue(sheet, fmt.Sprintf("C%d", currentRow), r.TransactionType)
		if r.Debit != 0 {
			f.SetCellValue(sheet, fmt.Sprintf("D%d", currentRow), formatRupiah(r.Debit))
		}
		if r.Credit != 0 {
			f.SetCellValue(sheet, fmt.Sprintf("E%d", currentRow), formatRupiah(r.Credit))
		}

		f.SetCellStyle(sheet, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("A%d", currentRow), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("B%d", currentRow), fmt.Sprintf("B%d", currentRow), styleCenter)
		f.SetCellStyle(sheet, fmt.Sprintf("C%d", currentRow), fmt.Sprintf("C%d", currentRow), styleLeft)
		f.SetCellStyle(sheet, fmt.Sprintf("D%d", currentRow), fmt.Sprintf("D%d", currentRow), styleRight)
		f.SetCellStyle(sheet, fmt.Sprintf("E%d", currentRow), fmt.Sprintf("E%d", currentRow), styleRight)
		currentRow++
	}

	// --- GRAND TOTAL ROW ---
	f.SetCellValue(sheet, fmt.Sprintf("A%d", currentRow), "TOTAL")
	f.MergeCell(sheet, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("C%d", currentRow))
	f.SetCellValue(sheet, fmt.Sprintf("D%d", currentRow), formatRupiah(totalDebit))
	f.SetCellValue(sheet, fmt.Sprintf("E%d", currentRow), formatRupiah(totalCredit))
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("E%d", currentRow), grandTotalStyle)

	currentRow += 2

	// --- SUMMARY SECTION ---
	f.SetCellValue(sheet, fmt.Sprintf("A%d", currentRow), "RINGKASAN AKHIR")
	f.MergeCell(sheet, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("D%d", currentRow))
	f.SetCellValue(sheet, fmt.Sprintf("E%d", currentRow), formatRupiah(totalDebit-totalCredit))

	summaryStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 12, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#2E7D32"}, Pattern: 1}, // Green for summary
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("E%d", currentRow), summaryStyle)
	f.SetRowHeight(sheet, currentRow, 25)

	f.SetColWidth(sheet, "A", "A", 5)
	f.SetColWidth(sheet, "B", "B", 30)
	f.SetColWidth(sheet, "C", "C", 20)
	f.SetColWidth(sheet, "D", "D", 20)
	f.SetColWidth(sheet, "E", "E", 20)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write excel: %w", err)
	}

	return buf.Bytes(), nil
}
