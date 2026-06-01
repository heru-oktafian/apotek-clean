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

func (s *ExportServices) ExportUnitConversionsToPDF(branchID string) ([]byte, error) {
	var conversions []models.UnitConversionDetail

	// Query dengan join untuk mendapatkan nama unit dan produk
	err := s.db.Table("unit_conversions").
		Select("unit_conversions.id, unit_conversions.value_conv, unit_conversions.branch_id, units_init.name as init_name, units_final.name as final_name, products.name as product_name").
		Joins("LEFT JOIN units AS units_init ON units_init.id = unit_conversions.init_id").
		Joins("LEFT JOIN units AS units_final ON units_final.id = unit_conversions.final_id").
		Joins("LEFT JOIN products ON products.id = unit_conversions.product_id").
		Where("unit_conversions.branch_id = ?", branchID).
		Order("unit_conversions.id ASC").
		Find(&conversions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch unit conversions: %w", err)
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
				text.New("UNIT CONVERSIONS", props.Text{
					Size:  14,
					Align: "center",
					Color: &props.Color{Red: 18, Green: 104, Blue: 202},
					Style: fontstyle.Bold,
				}),
			),
		),
	)

	headerRowContent := row.New(8).Add(
		col.New(3).WithStyle(headerCell()).Add(text.New("INIT UNIT", headerTextProps())),
		col.New(3).WithStyle(headerCell()).Add(text.New("FINAL UNIT", headerTextProps())),
		col.New(3).WithStyle(headerCell()).Add(text.New("VALUE CONV", headerTextProps())),
		col.New(3).WithStyle(headerCell()).Add(text.New("PRODUCT", headerTextProps())),
	)
	m.AddRows(headerRowContent)

	const rowsPerPageFirst = 20
	const rowsPerPageOther = 25

	rowCounter := 0
	isFirstPage := true

	for i, conv := range conversions {
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
				col.New(3).WithStyle(cellStyle).Add(text.New(conv.InitName, textProps)),
				col.New(3).WithStyle(cellStyle).Add(text.New(conv.FinalName, textProps)),
				col.New(3).WithStyle(cellStyle).Add(text.New(fmt.Sprintf("%d", conv.ValueConv), textProps)),
				col.New(3).WithStyle(cellStyle).Add(text.New(conv.ProductName, textProps)),
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
