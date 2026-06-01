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

func (s *ExportServices) ExportPurchaseItemsToPDF(branchID string, purchaseID string) ([]byte, error) {
	var items []models.AllPurchaseItems

	// Query data purchase items
	query := s.db.Table("purchase_items").
		Select("purchase_items.id, purchase_items.purchase_id, purchase_items.product_id, products.name as product_name, purchase_items.price, purchase_items.qty, purchase_items.unit_id, units.name as unit_name, purchase_items.sub_total, purchase_items.expired_date").
		Joins("JOIN products ON products.id = purchase_items.product_id").
		Joins("JOIN units ON units.id = purchase_items.unit_id").
		Where("purchase_items.purchase_id = ?", purchaseID)

	err := query.Order("products.name ASC").Find(&items).Error
	if err != nil {
		log.Printf("[ExportPurchaseItemsToPDF] Query error: %v", err)
		return nil, fmt.Errorf("failed to fetch purchase items: %w", err)
	}

	// Ambil header info
	var purchase models.AllPurchases
	// Query purchases dengan join ke suppliers
	err = s.db.Table("purchases").
		Select("purchases.id, purchases.supplier_id, suppliers.name as supplier_name, purchases.purchase_date, purchases.total_purchase, purchases.payment").
		Joins("JOIN suppliers ON suppliers.id = purchases.supplier_id").
		Where("purchases.id = ? AND purchases.branch_id = ?", purchaseID, branchID).
		First(&purchase).Error

	if err != nil {
		log.Printf("[ExportPurchaseItemsToPDF] Purchase not found or mismatch branch: %v", err)
		return nil, fmt.Errorf("purchase not found or access denied")
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
				text.New(fmt.Sprintf("ID PEMBELIAN : %s", purchase.ID), props.Text{
					Size:  14,
					Align: "center",
					Color: &props.Color{Red: 18, Green: 104, Blue: 202},
					Style: fontstyle.Bold,
				}),
			),
		),
		row.New(6).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("TANGGAL : %s", purchase.PurchaseDate.Format("02/01/2006")), props.Text{
					Size:  10,
					Align: "center",
				}),
			),
		),
		row.New(6).Add(
			col.New(12).Add(
				text.New(fmt.Sprintf("SUPPLIER : %s | METODE PEMBAYARAN : %s", purchase.SupplierName, purchase.Payment), props.Text{
					Size:  10,
					Align: "center",
				}),
			),
		),
	)

	// === TABLE HEADERS ===
	headerRowContent := row.New(8).Add(
		col.New(4).WithStyle(headerCell()).Add(text.New("PRODUK", headerTextProps())),
		col.New(2).WithStyle(headerCell()).Add(text.New("EXPIRED", headerTextProps())),
		col.New(2).WithStyle(headerCell()).Add(text.New("JUMLAH", headerTextProps())),
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

		// Override alignment for SubTotal
		subTotalProps := textProps
		subTotalProps.Align = "right"

		m.AddRows(
			row.New(8).Add(
				col.New(4).WithStyle(cellStyle).Add(text.New(item.ProductName, textProps)),
				col.New(2).WithStyle(cellStyle).Add(text.New(item.ExpiredDate.Format("02/01/2006"), textProps)),
				col.New(2).WithStyle(cellStyle).Add(text.New(fmt.Sprintf("%d %s", item.Qty, item.UnitName), textProps)),
				col.New(2).WithStyle(cellStyle).Add(text.New(formatRupiah(item.Price), textProps)),
				col.New(2).WithStyle(cellStyle).Add(text.New(formatRupiah(item.SubTotal), subTotalProps)),
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
			col.New(2).WithStyle(totalCellStyle).Add(text.New(formatRupiah(purchase.TotalPurchase), totalValueProps)),
		),
	)

	// Get PDF bytes
	document, err := m.Generate()
	if err != nil {
		log.Printf("[ExportPurchaseItemsToPDF] Generate error: %v", err)
		return nil, fmt.Errorf("failed to generate pdf: %w", err)
	}

	return document.GetBytes(), nil
}
