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

func (s *ExportServices) ExportFirstStockItemsToPDF(branchID string, firstStockID string) ([]byte, error) {
	var firstStockItems []models.AllFirstStockItems

	// Query data first stock items dengan join ke products dan units
	// Asumsi tabel products bernama 'products' dan units bernama 'units'
	// Sesuaikan nama tabel jika berbeda
	query := s.db.Table("first_stock_items").
		Select("first_stock_items.id, first_stock_items.first_stock_id, first_stock_items.product_id, products.name as product_name, first_stock_items.price, first_stock_items.qty, units.name as unit_name, first_stock_items.sub_total").
		Joins("JOIN products ON products.id = first_stock_items.product_id").
		Joins("JOIN units ON units.id = products.unit_id"). // Asumsi relasi unit ada di products
		Where("first_stock_items.first_stock_id = ?", firstStockID)

	err := query.Order("products.name ASC").Find(&firstStockItems).Error
	if err != nil {
		log.Printf("[ExportFirstStockItemsToPDF] Query error: %v", err)
		return nil, fmt.Errorf("failed to fetch first stock items: %w", err)
	}

	// Ambil header info (optional, misal deskripsi first stock)
	var firstStock models.FirstStocks
	if err := s.db.Where("id = ? AND branch_id = ?", firstStockID, branchID).First(&firstStock).Error; err != nil {
		// Jika tidak ketemu atau error, kita lanjut saja tanpa detail header,
		// atau return error jika strict. User request hanya bilang filter by first_stock_id.
		// Namun branch_id check penting untuk security
		log.Printf("[ExportFirstStockItemsToPDF] FirstStock not found or mismatch branch: %v", err)
		return nil, fmt.Errorf("first stock not found or access denied")
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
				text.New(fmt.Sprintf("STOK AWAL : %s", firstStock.ID), props.Text{
					Size:  14,
					Align: "center",
					Color: &props.Color{Red: 18, Green: 104, Blue: 202},
					Style: fontstyle.Bold,
				}),
			),
		),
		row.New(6).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("TANGGAL : %s | METODE PEMBAYARAN : %s", firstStock.FirstStockDate.Format("02/01/2006"), firstStock.Payment), props.Text{
					Size:  10,
					Align: "center",
				}),
			),
		),
		row.New(6).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("DESKRIPSI : %s", firstStock.Description), props.Text{
					Size:  10,
					Align: "center",
				}),
			),
		),
	)

	// === TABLE HEADERS ===
	headerRowContent := row.New(8).Add(
		col.New(6).WithStyle(headerCell()).Add(text.New("PRODUK", headerTextProps())),
		col.New(2).WithStyle(headerCell()).Add(text.New("QTY", headerTextProps())),
		col.New(2).WithStyle(headerCell()).Add(text.New("HARGA", headerTextProps())),
		col.New(2).WithStyle(headerCell()).Add(text.New("SUB TOTAL", headerTextProps())),
	)
	m.AddRows(headerRowContent)

	// === TABLE DATA ROWS ===
	const rowsPerPageFirst = 18 // Sesuaikan karena ada header tambahan
	const rowsPerPageOther = 22

	rowCounter := 0
	isFirstPage := true

	for i, fsi := range firstStockItems {
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
				col.New(6).WithStyle(cellStyle).Add(text.New(fsi.ProductName, textProps)),
				col.New(2).WithStyle(cellStyle).Add(text.New(fmt.Sprintf("%d %s", fsi.Qty, fsi.UnitName), textProps)),
				col.New(2).WithStyle(cellStyle).Add(text.New(formatRupiah(fsi.Price), textProps)),
				col.New(2).WithStyle(cellStyle).Add(text.New(formatRupiah(fsi.SubTotal), textProps)),
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
			col.New(2).WithStyle(totalCellStyle).Add(text.New(formatRupiah(firstStock.TotalFirstStock), totalValueProps)),
		),
	)

	// Get PDF bytes
	document, err := m.Generate()
	if err != nil {
		log.Printf("[ExportFirstStockItemsToPDF] Generate error: %v", err)
		return nil, fmt.Errorf("failed to generate pdf: %w", err)
	}

	return document.GetBytes(), nil
}
