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

func (s *ExportServices) ExportSaleReturnItemsToPDF(branchID string, saleReturnID string) ([]byte, error) {
	var items []models.AllSaleReturnItems

	// Query data sale_return_items
	// Note: Perlu join products -> units untuk unit_name.
	// Model AllSaleReturnItems menggunakan field ProId, ProName, UnitId, UnitName
	query := s.db.Table("sale_return_items").
		Select("sale_return_items.id, sale_return_items.sale_return_id, sale_return_items.product_id as pro_id, products.name as pro_name, sale_return_items.price, sale_return_items.qty, units.id as unit_id, units.name as unit_name, sale_return_items.sub_total, sale_return_items.expired_date").
		Joins("JOIN products ON products.id = sale_return_items.product_id").
		Joins("JOIN units ON units.id = products.unit_id").
		Where("sale_return_items.sale_return_id = ?", saleReturnID)

	err := query.Order("products.name ASC").Find(&items).Error
	if err != nil {
		log.Printf("[ExportSaleReturnItemsToPDF] Query error: %v", err)
		return nil, fmt.Errorf("failed to fetch sale return items: %w", err)
	}

	// Ambil header info
	var saleReturn models.SaleReturns
	if err := s.db.Where("id = ? AND branch_id = ?", saleReturnID, branchID).First(&saleReturn).Error; err != nil {
		log.Printf("[ExportSaleReturnItemsToPDF] SaleReturn not found or mismatch branch: %v", err)
		return nil, fmt.Errorf("sale return not found or access denied")
	}

	// Konfigurasi PDF
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
				text.New(fmt.Sprintf("ID RETUR PENJUALAN : %s", saleReturn.ID), props.Text{
					Size:  14,
					Align: "center",
					Color: &props.Color{Red: 18, Green: 104, Blue: 202},
					Style: fontstyle.Bold,
				}),
			),
		),
		row.New(6).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("TANGGAL RETUR : %s", saleReturn.ReturnDate.Format("02/01/2006")), props.Text{
					Size:  10,
					Align: "center",
				}),
			),
		),
		row.New(6).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("ID PENJUALAN : %s | METODE PEMBAYARAN : %s", saleReturn.SaleId, saleReturn.Payment), props.Text{
					Size:  10,
					Align: "center",
				}),
			),
		),
	)

	// === TABLE HEADERS ===
	headerRowContent := row.New(8).Add(
		col.New(4).WithStyle(headerCell()).Add(text.New("PRODUK", headerTextProps())),
		col.New(2).WithStyle(headerCell()).Add(text.New("KADALURSA", headerTextProps())),
		col.New(2).WithStyle(headerCell()).Add(text.New("HARGA", headerTextProps())),
		col.New(2).WithStyle(headerCell()).Add(text.New("QTY", headerTextProps())),
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
				col.New(4).WithStyle(cellStyle).Add(text.New(item.ProName, textProps)),
				col.New(2).WithStyle(cellStyle).Add(text.New(item.ExpiredDate.Format("02/01/2006"), textProps)),
				col.New(2).WithStyle(cellStyle).Add(text.New(formatRupiah(item.Price), textProps)),
				col.New(2).WithStyle(cellStyle).Add(text.New(fmt.Sprintf("%d %s", item.Qty, item.UnitName), textProps)),
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
			col.New(10).WithStyle(totalCellStyle).Add(text.New("TOTAL", totalTextProps)),
			col.New(2).WithStyle(totalCellStyle).Add(text.New(formatRupiah(saleReturn.TotalReturn), totalValueProps)),
		),
	)

	// Get PDF bytes
	document, err := m.Generate()
	if err != nil {
		log.Printf("[ExportSaleReturnItemsToPDF] Generate error: %v", err)
		return nil, fmt.Errorf("failed to generate pdf: %w", err)
	}

	return document.GetBytes(), nil
}
