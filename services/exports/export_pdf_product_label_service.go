package services

import (
	"fmt"

	"github.com/heru-oktafian/fiber-apotek/models"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/code"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

// ExportProductLabelToPDF generates product labels with barcode, name, and price
func (s *ExportServices) ExportProductLabelToPDF(productID string, branchID string, qty int) ([]byte, error) {
	var product models.ProductDetail

	err := s.db.Model(&models.Product{}).
		Select(`
			products.id,
			products.sku,
			products.name,
			products.sales_price
		`).
		Where("products.id = ? AND products.branch_id = ?", productID, branchID).
		First(&product).Error

	if err != nil {
		return nil, fmt.Errorf("produk tidak ditemukan: %w", err)
	}

	if qty <= 0 {
		qty = 1
	}

	// Konfigurasi PDF untuk label (ukuran kecil, misal 50mm x 30mm)
	// Kita gunakan ukuran kustom agar pas dengan printer thermal label
	cfg := config.NewBuilder().
		WithDimensions(50, 30).
		WithLeftMargin(2).
		WithTopMargin(2).
		WithRightMargin(2).
		WithBottomMargin(2).
		Build()

	m := maroto.New(cfg)

	for i := 0; i < qty; i++ {
		// Setiap label di halaman baru (kecuali yang pertama yang otomatis)
		if i > 0 {
			// Maroto v2 automatically adds pages on overflow,
			// but for specific label sizes, we might need a manual break if it doesn't fit exactly.
		}
		// Nama Produk - dikurangi tingginya untuk mendekatkan ke barcode
		m.AddRows(
			row.New(5).Add(
				col.New(12).Add(
					text.New(product.Name, props.Text{
						Size:  8,
						Align: "center",
						Style: fontstyle.Bold,
					}),
				),
			),
		)

		// Barcode (SKU)
		m.AddRows(
			row.New(12).Add(
				col.New(12).Add(
					code.NewBar(product.SKU, props.Barcode{
						Percent: 95,
						Center:  true,
					}),
				),
			),
		)

		// SKU (Kiri) & Harga (Kanan) - Sejajar dalam satu baris
		m.AddRows(
			row.New(6).Add(
				col.New(6).Add(
					text.New(product.SKU, props.Text{
						Size:  8,
						Align: "left",
					}),
				),
				col.New(6).Add(
					text.New(formatRupiah(product.SalesPrice), props.Text{
						Size:  8, // Ukuran disesuaikan agar muat
						Align: "right",
						Style: fontstyle.Bold,
					}),
				),
			),
		)
	}

	// Generate PDF
	document, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("gagal generate label pdf: %w", err)
	}

	return document.GetBytes(), nil
}
