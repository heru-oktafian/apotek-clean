package services

import (
	"fmt"
	"log"
	"time"

	"github.com/heru-oktafian/fiber-apotek/models"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

func (s *ExportServices) ExportBuyReturnsToPDF(branchID string, month string) ([]byte, error) {
	var buyReturns []models.BuyReturns

	query := s.db.Where("branch_id = ?", branchID)

	// Filter by month if provided (format: YYYY-MM)
	if month != "" {
		parsedTime, err := time.Parse("2006-01", month)
		if err == nil {
			startDate := parsedTime
			endDate := parsedTime.AddDate(0, 1, 0)
			query = query.Where("return_date >= ? AND return_date < ?", startDate, endDate)
		}
	}

	err := query.Order("return_date DESC").Find(&buyReturns).Error
	if err != nil {
		log.Printf("[ExportBuyReturnsToPDF] Query error: %v", err)
		return nil, fmt.Errorf("failed to fetch buy returns: %w", err)
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
				text.New(fmt.Sprintf("RETUR PEMBELIAN %s", month), props.Text{
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
		col.New(6).WithStyle(headerCell()).Add(text.New("PURCHASE ID", headerTextProps())),
		col.New(2).WithStyle(headerCell()).Add(text.New("TANGGAL", headerTextProps())),
		col.New(2).WithStyle(headerCell()).Add(text.New("TOTAL", headerTextProps())),
	)
	m.AddRows(headerRowContent)

	// === TABLE DATA ROWS ===
	const rowsPerPageFirst = 21 // Baris per halaman untuk halaman pertama
	const rowsPerPageOther = 22 // Baris per halaman untuk halaman lainnya

	rowCounter := 0
	isFirstPage := true

	// Hitung total dari semua TotalReturn
	var grandTotal int
	for _, br := range buyReturns {
		grandTotal += br.TotalReturn
	}

	for i, br := range buyReturns {
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

		m.AddRows(
			row.New(8).Add(
				col.New(2).WithStyle(cellStyle).Add(text.New(br.ID, textProps)),
				col.New(6).WithStyle(cellStyle).Add(text.New(br.PurchaseId, textProps)),
				col.New(2).WithStyle(cellStyle).Add(text.New(br.ReturnDate.Format("02/01/2006"), textProps)),
				col.New(2).WithStyle(cellStyle).Add(text.New(formatRupiah(br.TotalReturn), textProps)),
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
		Align: "center",
	}
	totalValueProps := props.Text{
		Size:  10,
		Style: fontstyle.Bold,
		Color: &props.Color{Red: 255, Green: 255, Blue: 255}, // Putih
		Align: "right",
	}

	m.AddRows(
		row.New(8).Add(
			col.New(10).WithStyle(totalCellStyle).Add(text.New("TOTAL", totalTextProps)),
			col.New(2).WithStyle(totalCellStyle).Add(text.New(formatRupiah(grandTotal), totalValueProps)),
		),
	)

	// Get PDF bytes
	document, err := m.Generate()
	if err != nil {
		log.Printf("[ExportBuyReturnsToPDF] Generate error: %v", err)
		return nil, fmt.Errorf("failed to generate pdf: %w", err)
	}

	return document.GetBytes(), nil
}
