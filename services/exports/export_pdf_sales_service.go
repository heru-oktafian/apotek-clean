package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/heru-oktafian/fiber-apotek/models"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

func (s *ExportServices) ExportSalesToPDF(branchID string, month string) ([]byte, error) {
	var sales []models.Sales

	query := s.db.Where("branch_id = ?", branchID)

	// Filter by month if provided (format: YYYY-MM)
	if month != "" {
		parsedTime, err := time.Parse("2006-01", month)
		if err == nil {
			startDate := parsedTime
			endDate := parsedTime.AddDate(0, 1, 0)
			query = query.Where("sale_date >= ? AND sale_date < ?", startDate, endDate)
		}
	}

	err := query.Order("sale_date DESC").Find(&sales).Error
	if err != nil {
		log.Printf("[ExportSalesToPDF] Query error: %v", err)
		return nil, fmt.Errorf("failed to fetch sales: %w", err)
	}

	// Konfigurasi PDF dengan landscape orientation
	cfg := config.NewBuilder().
		WithPageNumber().
		WithOrientation(orientation.Horizontal).
		WithLeftMargin(10).
		WithTopMargin(15).
		WithRightMargin(10).
		WithBottomMargin(10).
		Build()

	m := maroto.New(cfg)

	// === HEADER JUDUL ===
	m.AddRows(
		row.New(9).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("PENJUALAN %s", month), props.Text{
					Size:  14,
					Align: "center",
					Color: &props.Color{Red: 18, Green: 104, Blue: 202},
					Style: fontstyle.Bold,
				}),
			),
		),
	)

	// === TABLE HEADERS ===
	headerRowContent := row.New(8).Add(
		col.New(2).WithStyle(headerCell()).Add(text.New("ID", headerTextProps())),
		col.New(4).WithStyle(headerCell()).Add(text.New("KETERANGAN", headerTextProps())),
		col.New(2).WithStyle(headerCell()).Add(text.New("TANGGAL", headerTextProps())),
		col.New(2).WithStyle(headerCell()).Add(text.New("PEMBAYARAN", headerTextProps())),
		col.New(1).WithStyle(headerCell()).Add(text.New("SALES", headerTextProps())),
		col.New(1).WithStyle(headerCell()).Add(text.New("MARGIN", headerTextProps())),
	)
	m.AddRows(headerRowContent)

	// === TABLE DATA ROWS ===
	const rowsPerPageFirst = 21 // Baris per halaman untuk halaman pertama
	const rowsPerPageOther = 22 // Baris per halaman untuk halaman lainnya

	rowCounter := 0
	isFirstPage := true

	// Hitung total dari semua TotalSale dan Margin
	var grandTotal int
	var grandMargin int
	for _, sale := range sales {
		grandTotal += sale.TotalSale
		grandMargin += sale.ProfitEstimate
	}

	for i, sale := range sales {
		var maxRowsThisPage int
		if isFirstPage {
			maxRowsThisPage = rowsPerPageFirst
		} else {
			maxRowsThisPage = rowsPerPageOther
		}

		// Tambah header baru jika sudah mencapai batas halaman
		if rowCounter > 0 && rowCounter >= maxRowsThisPage {
			m.AddRows(headerRowContent)
			rowCounter = 0
			isFirstPage = false
		}

		// Logika KETERANGAN: Ambil nama item
		var itemNames []string
		if err := s.db.Table("sale_items sit").
			Select("pro.name").
			Joins("LEFT JOIN products pro ON pro.id = sit.product_id").
			Where("sit.sale_id = ?", sale.ID).
			Order("pro.name ASC").
			Pluck("pro.name", &itemNames).Error; err != nil {
			log.Printf("[ExportSalesToPDF] Failed to fetch item names for %s: %v", sale.ID, err)
		}

		descItems := strings.Join(itemNames, ", ")
		dateWith7 := sale.SaleDate.Add(7 * time.Hour).Format("02-01-2006 15:04")
		var description string
		if descItems != "" {
			description = descItems + " ; " + dateWith7
		} else {
			description = dateWith7
		}

		var cellStyle *props.Cell
		var textProps props.Text

		// Alternating row colors
		if i%2 == 0 {
			cellStyle = dataCellWhite()
			textProps = dataPropsWhite()
		} else {
			cellStyle = dataCellGray()
			textProps = dataPropsGray()
		}

		textPropsLeft := textProps
		textPropsLeft.Align = align.Left
		textPropsCenter := textProps
		textPropsCenter.Align = align.Center
		textPropsRight := textProps
		textPropsRight.Align = align.Right

		m.AddRows(
			row.New(8).Add(
				col.New(2).WithStyle(cellStyle).Add(text.New(sale.ID, textPropsCenter)),
				col.New(4).WithStyle(cellStyle).Add(text.New(description, textPropsCenter)),
				col.New(2).WithStyle(cellStyle).Add(text.New(sale.SaleDate.Format("02/01/2006"), textPropsCenter)),
				col.New(2).WithStyle(cellStyle).Add(text.New(string(sale.Payment), textPropsCenter)),
				col.New(1).WithStyle(cellStyle).Add(text.New(formatRupiah(sale.TotalSale), textPropsCenter)),
				col.New(1).WithStyle(cellStyle).Add(text.New(formatRupiah(sale.ProfitEstimate), textPropsCenter)),
			),
		)

		rowCounter++
	}

	// === BARIS TOTAL ===
	// Style untuk baris total (background biru, text putih, bold)
	totalCellStyle := &props.Cell{
		BackgroundColor: &props.Color{Red: 18, Green: 104, Blue: 202},
	}
	totalTextProps := props.Text{
		Size:  10,
		Style: fontstyle.Bold,
		Color: &props.Color{Red: 255, Green: 255, Blue: 255}, // Putih
		Align: align.Center,
	}
	totalValueProps := props.Text{
		Size:  10,
		Style: fontstyle.Bold,
		Color: &props.Color{Red: 255, Green: 255, Blue: 255}, // Putih
		Align: align.Center,
	}

	m.AddRows(
		row.New(8).Add(
			col.New(10).WithStyle(totalCellStyle).Add(text.New("GRAND TOTAL", totalTextProps)),
			col.New(1).WithStyle(totalCellStyle).Add(text.New(formatRupiah(grandTotal), totalValueProps)),
			col.New(1).WithStyle(totalCellStyle).Add(text.New(formatRupiah(grandMargin), totalValueProps)),
		),
	)

	// Get PDF bytes
	document, err := m.Generate()
	if err != nil {
		log.Printf("[ExportSalesToPDF] Generate error: %v", err)
		return nil, fmt.Errorf("failed to generate pdf: %w", err)
	}

	return document.GetBytes(), nil
}
