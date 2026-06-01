package services

import (
	"fmt"

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

func (s *ExportServices) ExportUnitsToPDF(branchID string) ([]byte, error) {
	var units []models.Unit

	err := s.db.Where("branch_id = ?", branchID).Order("name ASC").Find(&units).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch units: %w", err)
	}

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
				text.New("MASTER UNITS", props.Text{
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
		col.New(4).WithStyle(headerCell()).Add(text.New("UNIT ID", headerTextProps())),
		col.New(8).WithStyle(headerCell()).Add(text.New("UNIT NAME", headerTextProps())),
	)
	m.AddRows(headerRowContent)

	// === TABLE DATA ROWS ===
	const rowsPerPageFirst = 20
	const rowsPerPageOther = 25

	rowCounter := 0
	isFirstPage := true

	for i, u := range units {
		var maxRowsThisPage int
		if isFirstPage {
			maxRowsThisPage = rowsPerPageFirst
		} else {
			maxRowsThisPage = rowsPerPageOther
		}

		if rowCounter > 0 && rowCounter >= maxRowsThisPage {
			m.AddRows(headerRowContent)
			rowCounter = 0
			isFirstPage = false
		}

		var cellStyle *props.Cell
		var textProps props.Text

		if i%2 == 0 {
			cellStyle = dataCellWhite()
			textProps = dataPropsWhite()
		} else {
			cellStyle = dataCellGray()
			textProps = dataPropsGray()
		}

		m.AddRows(
			row.New(8).Add(
				col.New(4).WithStyle(cellStyle).Add(text.New(u.ID, textProps)),
				col.New(8).WithStyle(cellStyle).Add(text.New(u.Name, textProps)),
			),
		)
		rowCounter++
	}

	document, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate pdf: %w", err)
	}

	return document.GetBytes(), nil
}
