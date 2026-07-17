package services

import (
	"fmt"
	"log"
	"time"

	"apotek-clean/internal/core/entities"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

func (s *ExportServices) ExportOpnamesToPDF(branchID string, month string) ([]byte, error) {
	var opnames []models.Opnames

	query := s.db.Where("branch_id = ?", branchID)

	// Filter by month if provided (format: YYYY-MM)
	if month != "" {
		parsedTime, err := time.Parse("2006-01", month)
		if err == nil {
			startDate := parsedTime
			endDate := parsedTime.AddDate(0, 1, 0)
			query = query.Where("opname_date >= ? AND opname_date < ?", startDate, endDate)
		}
	}

	err := query.Order("opname_date DESC").Find(&opnames).Error
	if err != nil {
		log.Printf("[ExportOpnamesToPDF] Query error: %v", err)
		return nil, fmt.Errorf("failed to fetch opnames: %w", err)
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
				text.New("MASTER OPNAMES", props.Text{
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
		col.New(3).WithStyle(headerCell()).Add(text.New("DESCRIPTION", headerTextProps())),
		col.New(2).WithStyle(headerCell()).Add(text.New("DATE", headerTextProps())),
		col.New(2).WithStyle(headerCell()).Add(text.New("STATUS", headerTextProps())),
		col.New(2).WithStyle(headerCell()).Add(text.New("TOTAL", headerTextProps())),
	)
	m.AddRows(headerRowContent)

	// === TABLE DATA ROWS ===
	const rowsPerPageFirst = 21 // Baris per halaman untuk halaman pertama
	const rowsPerPageOther = 22 // Baris per halaman untuk halaman lainnya

	rowCounter := 0
	isFirstPage := true

	for i, opname := range opnames {
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
				col.New(2).WithStyle(cellStyle).Add(text.New(opname.ID, textProps)),
				col.New(3).WithStyle(cellStyle).Add(text.New(opname.Description, textProps)),
				col.New(2).WithStyle(cellStyle).Add(text.New(opname.OpnameDate.Format("02/01/2006"), textProps)),
				col.New(2).WithStyle(cellStyle).Add(text.New(string(opname.OpnameStatus), textProps)),
				col.New(2).WithStyle(cellStyle).Add(text.New(formatRupiah(opname.TotalOpname), textProps)),
			),
		)

		rowCounter++
	}

	// Get PDF bytes
	document, err := m.Generate()
	if err != nil {
		log.Printf("[ExportOpnamesToPDF] Generate error: %v", err)
		return nil, fmt.Errorf("failed to generate pdf: %w", err)
	}

	return document.GetBytes(), nil
}
