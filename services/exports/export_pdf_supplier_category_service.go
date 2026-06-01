package services

import (
	"fmt"
	"log"

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

func (s *ExportServices) ExportSupplierCategoriesToPDF(branchID string) ([]byte, error) {
	var categories []models.SupplierCategory

	err := s.db.Where("branch_id = ?", branchID).Order("name ASC").Find(&categories).Error
	if err != nil {
		log.Printf("[ExportSupplierCategoriesToPDF] Query error: %v", err)
		return nil, fmt.Errorf("failed to fetch supplier categories: %w", err)
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
				text.New("MASTER SUPPLIER CATEGORIES", props.Text{
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
		col.New(6).WithStyle(headerCell()).Add(text.New("CATEGORY ID", headerTextProps())),
		col.New(6).WithStyle(headerCell()).Add(text.New("CATEGORY NAME", headerTextProps())),
	)
	m.AddRows(headerRowContent)

	// === TABLE DATA ROWS ===
	const rowsPerPageFirst = 21 // Baris per halaman untuk halaman pertama
	const rowsPerPageOther = 22 // Baris per halaman untuk halaman lainnya

	rowCounter := 0
	isFirstPage := true

	for i, cat := range categories {
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
				col.New(6).WithStyle(cellStyle).Add(text.New(fmt.Sprintf("%d", cat.ID), textProps)),
				col.New(6).WithStyle(cellStyle).Add(text.New(cat.Name, textProps)),
			),
		)

		rowCounter++
	}

	// Generate PDF
	document, err := m.Generate()
	if err != nil {
		log.Printf("[ExportSupplierCategoriesToPDF] Generate error: %v", err)
		return nil, fmt.Errorf("failed to generate pdf: %w", err)
	}

	return document.GetBytes(), nil
}
