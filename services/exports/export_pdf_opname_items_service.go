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

type OpnameItemPDF struct {
	ProductName string
	UnitName    string
	QtySystem   int
	QtyPhysical int
	QtyDiff     int
	Price       int
	SubTotal    int
}

func (s *ExportServices) ExportOpnameItemsToPDF(branchID string, opnameID string) ([]byte, error) {
	var items []OpnameItemPDF

	// Query data opname items dengan join ke products dan units
	// Menghitung selisih (diff) antara Qty (Fisik) dan QtyExist (Sistem)
	query := s.db.Table("opname_items").
		Select("products.name as product_name, units.name as unit_name, opname_items.qty_exist as qty_system, opname_items.qty as qty_physical, (opname_items.qty - opname_items.qty_exist) as qty_diff, opname_items.price, opname_items.sub_total").
		Joins("JOIN products ON products.id = opname_items.product_id").
		Joins("JOIN units ON units.id = products.unit_id").
		Where("opname_items.opname_id = ?", opnameID)

	err := query.Order("products.name ASC").Scan(&items).Error
	if err != nil {
		log.Printf("[ExportOpnameItemsToPDF] Query error: %v", err)
		return nil, fmt.Errorf("failed to fetch opname items: %w", err)
	}

	// Ambil header info (opname detail)
	var opname models.Opnames
	if err := s.db.Where("id = ? AND branch_id = ?", opnameID, branchID).First(&opname).Error; err != nil {
		log.Printf("[ExportOpnameItemsToPDF] Opname not found or mismatch branch: %v", err)
		return nil, fmt.Errorf("opname not found or access denied")
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
				text.New(fmt.Sprintf("OPNAME : %s", opname.ID), props.Text{
					Size:  14,
					Align: "center",
					Color: &props.Color{Red: 18, Green: 104, Blue: 202},
					Style: fontstyle.Bold,
				}),
			),
		),
		row.New(6).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("TANGGAL : %s | METODE PEMBAYARAN : %s", opname.OpnameDate.Format("02/01/2006"), opname.Payment), props.Text{
					Size:  10,
					Align: "center",
				}),
			),
		),
		row.New(6).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("KETERANGAN : %s", opname.Description), props.Text{
					Size:  10,
					Align: "center",
				}),
			),
		),
	)

	// === TABLE HEADERS ===
	headerRowContent := row.New(8).Add(
		col.New(5).WithStyle(headerCell()).Add(text.New("PRODUK", headerTextProps())),
		col.New(1).WithStyle(headerCell()).Add(text.New("SYSTEM", headerTextProps())),
		col.New(1).WithStyle(headerCell()).Add(text.New("FISIK", headerTextProps())),
		col.New(1).WithStyle(headerCell()).Add(text.New("SELISIH", headerTextProps())),
		col.New(2).WithStyle(headerCell()).Add(text.New("HARGA", headerTextProps())),
		col.New(2).WithStyle(headerCell()).Add(text.New("SUB TOTAL", headerTextProps())),
	)
	m.AddRows(headerRowContent)

	// === TABLE DATA ROWS ===
	const rowsPerPageFirst = 18
	const rowsPerPageOther = 22

	rowCounter := 0
	isFirstPage := true

	for i, item := range items {
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
				col.New(5).WithStyle(cellStyle).Add(text.New(item.ProductName, textProps)),
				col.New(1).WithStyle(cellStyle).Add(text.New(fmt.Sprintf("%d %s", item.QtySystem, item.UnitName), textProps)),
				col.New(1).WithStyle(cellStyle).Add(text.New(fmt.Sprintf("%d %s", item.QtyPhysical, item.UnitName), textProps)),
				col.New(1).WithStyle(cellStyle).Add(text.New(fmt.Sprintf("%d %s", item.QtyDiff, item.UnitName), textProps)),
				col.New(2).WithStyle(cellStyle).Add(text.New(formatRupiah(item.Price), textProps)),
				col.New(2).WithStyle(cellStyle).Add(text.New(formatRupiah(item.SubTotal), textProps)),
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
			col.New(10).WithStyle(totalCellStyle).Add(text.New("TOTAL SELISIH", totalTextProps)),
			col.New(2).WithStyle(totalCellStyle).Add(text.New(formatRupiah(opname.TotalOpname), totalValueProps)),
		),
	)

	// Get PDF bytes
	document, err := m.Generate()
	if err != nil {
		log.Printf("[ExportOpnameItemsToPDF] Generate error: %v", err)
		return nil, fmt.Errorf("failed to generate pdf: %w", err)
	}

	return document.GetBytes(), nil
}
