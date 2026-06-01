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

func (s *ExportServices) ExportMemberCategoriesToPDF(branchID string) ([]byte, error) {
	var categories []models.MemberCategory

	err := s.db.Where("branch_id = ?", branchID).Order("name ASC").Find(&categories).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch member categories: %w", err)
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

	m.AddRows(
		row.New(9).Add(
			col.New(12).Add(
				text.New("MEMBER CATEGORIES", props.Text{
					Size:  14,
					Align: "center",
					Color: &props.Color{Red: 18, Green: 104, Blue: 202},
					Style: fontstyle.Bold,
				}),
			),
		),
	)

	headerRowContent := row.New(8).Add(
		col.New(4).WithStyle(headerCell()).Add(text.New("CATEGORY ID", headerTextProps())),
		col.New(8).WithStyle(headerCell()).Add(text.New("CATEGORY NAME", headerTextProps())),
	)
	m.AddRows(headerRowContent)

	const rowsPerPageFirst = 20
	const rowsPerPageOther = 25

	rowCounter := 0
	isFirstPage := true

	for i, cat := range categories {
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
				col.New(4).WithStyle(cellStyle).Add(text.New(fmt.Sprintf("%d", cat.ID), textProps)),
				col.New(8).WithStyle(cellStyle).Add(text.New(cat.Name, textProps)),
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
