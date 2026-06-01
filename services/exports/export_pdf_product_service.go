package services

import (
	"fmt"

	"github.com/heru-oktafian/fiber-apotek/models"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

// ExportProductsToPDF — FIXED: Tabel rapi, garis, header biru, alignment left
func (s *ExportServices) ExportProductsToPDF(branchID string) ([]byte, error) {
	var products []models.ProductDetail

	err := s.db.Model(&models.Product{}).
		Select(`
			products.id,
			products.sku,
			products.name,
			products.alias,
			products.purchase_price,
			products.sales_price,
			products.alternate_price,
			products.stock,
			products.expired_date,
			units.name as unit_name,
			product_categories.name as product_category_name
		`).
		Joins("LEFT JOIN units ON units.id = products.unit_id").
		Joins("LEFT JOIN product_categories ON product_categories.id = products.product_category_id").
		Where("products.branch_id = ?", branchID).
		Order("products.name ASC").
		Find(&products).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", err)
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
				text.New("MASTER PRODUCTS", props.Text{
					Size:  14,
					Align: "center",
					Color: &props.Color{Red: 18, Green: 104, Blue: 202},
					Style: fontstyle.Bold,
				}),
			),
		),
	)

	// === TABLE HEADERS (akan berulang di setiap halaman) ===
	headerRowContent := row.New(8).Add(
		col.New(2).WithStyle(headerCell()).Add(text.New("SKU", headerTextProps())),
		col.New(2).WithStyle(headerCell()).Add(text.New("NAME", headerTextProps())),
		col.New(2).WithStyle(headerCell()).Add(text.New("ALIAS", headerTextProps())),
		col.New(1).WithStyle(headerCell()).Add(text.New("PURCHASE", headerTextProps())),
		col.New(1).WithStyle(headerCell()).Add(text.New("SALE", headerTextProps())),
		col.New(1).WithStyle(headerCell()).Add(text.New("ALTERNATIF", headerTextProps())),
		col.New(1).WithStyle(headerCell()).Add(text.New("STOCK", headerTextProps())),
		col.New(1).WithStyle(headerCell()).Add(text.New("UNIT", headerTextProps())),
		col.New(1).WithStyle(headerCell()).Add(text.New("EXPIRED", headerTextProps())),
	)
	m.AddRows(headerRowContent)

	// === TABLE DATA ROWS ===
	// Setiap row = 8mm height
	// Halaman pertama: ~168mm tersedia - 9mm (header judul) - 8mm (table header) = 151mm = max 18 rows
	// Halaman berikutnya: ~268mm - 8mm (table header) = 260mm = max 32 rows
	// Kita gunakan 20 untuk halaman pertama dan 25 untuk halaman berikutnya sebagai safe margin

	const rowsPerPageFirst = 21 // Baris per halaman untuk halaman pertama
	const rowsPerPageOther = 22 // Baris per halaman untuk halaman lainnya

	rowCounter := 0
	isFirstPage := true

	for i, p := range products {
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
				col.New(2).WithStyle(cellStyle).Add(text.New(p.SKU, textProps)),
				col.New(2).WithStyle(cellStyle).Add(text.New(p.Name, textProps)),
				col.New(2).WithStyle(cellStyle).Add(text.New(p.Alias, textProps)),
				col.New(1).WithStyle(cellStyle).Add(text.New(formatRupiah(p.PurchasePrice), textProps)),
				col.New(1).WithStyle(cellStyle).Add(text.New(formatRupiah(p.SalesPrice), textProps)),
				col.New(1).WithStyle(cellStyle).Add(text.New(formatRupiah(p.AlternatePrice), textProps)),
				col.New(1).WithStyle(cellStyle).Add(text.New(fmt.Sprintf("%d", p.Stock), textProps)),
				col.New(1).WithStyle(cellStyle).Add(text.New(p.UnitName, textProps)),
				col.New(1).WithStyle(cellStyle).Add(text.New(p.ExpiredDate.Format("02/01/2006"), textProps)),
			),
		)
		rowCounter++
	}

	// Generate PDF
	document, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate pdf: %w", err)
	}

	return document.GetBytes(), nil
}

// headerTextProps — styling untuk text header table (putih, bold, center)
func headerTextProps() props.Text {
	return props.Text{
		Size:  9,
		Align: "center",
		Color: &props.Color{Red: 255, Green: 255, Blue: 255}, // Putih
		Style: fontstyle.Bold,
	}
}

// dataPropsWhite — styling untuk data rows dengan background putih
func dataPropsWhite() props.Text {
	return props.Text{
		Size:  8,
		Align: "left",
		Color: &props.Color{Red: 0, Green: 0, Blue: 0}, // Hitam
	}
}

// dataPropsGray — styling untuk data rows dengan background abu-abu
func dataPropsGray() props.Text {
	return props.Text{
		Size:  8,
		Align: "left",
		Color: &props.Color{Red: 0, Green: 0, Blue: 0}, // Hitam
	}
}

// headerCell — cell properties untuk header dengan background biru dan border
func headerCell() *props.Cell {
	return &props.Cell{
		BackgroundColor: &props.Color{Red: 0, Green: 102, Blue: 204}, // Biru
		BorderType:      border.Full,
		BorderThickness: 0.2,
		BorderColor:     &props.Color{Red: 0, Green: 0, Blue: 0},
	}
}

// dataCellWhite — cell properties untuk data rows putih dengan border
func dataCellWhite() *props.Cell {
	return &props.Cell{
		BackgroundColor: &props.Color{Red: 255, Green: 255, Blue: 255}, // Putih
		BorderType:      border.Full,
		BorderThickness: 0.2,
		BorderColor:     &props.Color{Red: 192, Green: 192, Blue: 192}, // Abu-abu terang
	}
}

// dataCellGray — cell properties untuk data rows abu-abu dengan border
func dataCellGray() *props.Cell {
	return &props.Cell{
		BackgroundColor: &props.Color{Red: 240, Green: 240, Blue: 240}, // Abu-abu terang
		BorderType:      border.Full,
		BorderThickness: 0.2,
		BorderColor:     &props.Color{Red: 192, Green: 192, Blue: 192}, // Abu-abu terang
	}
}
