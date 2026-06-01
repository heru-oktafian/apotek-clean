package seeders

import (
	config "github.com/heru-oktafian/fiber-apotek/configs"
	models "github.com/heru-oktafian/fiber-apotek/models"
)

// SupplierSeed function untuk menambahkan data supplier ke database
func SupplierSeed() {
	supplier := []models.Supplier{
		// {ID: "SPL250207144602", Name: "PT. Lab Medika Sejahtera", Phone: "(031)7403777, (031)7458102", Address: "Jl. Raya Manukan Kulon 60, Blok E, No. 2 - Surabaya", PIC: "Tutut Siswoyo (082329289326)", SupplierCategoryId: 3, BranchID: "BRC250118132203"},
		// {ID: "SPL250207144603", Name: "PT. Lestari Jaya Farma", Phone: "(0354)7418787", Address: "Banjaran Gg II, No. 44 - Kediri", PIC: "Manda (085606946466)", SupplierCategoryId: 3, BranchID: "BRC250118132203"},
		// {ID: "SPL250207144604", Name: "PT. Gelora Gempita Farma", Phone: "(031)99681266", Address: "Jl. Raya Tropodo, Blok Y, No. 12A, Tropodo, Waru - Sidoarjo", PIC: "Kusnaini (081330332013)", SupplierCategoryId: 3, BranchID: "BRC250118132203"},
		// {ID: "SPL250207144605", Name: "PT. Sapta Sari Tama", Phone: "-", Address: "Jl. Dukuh Pakis, No. 11 - Surabaya", PIC: "Mia (08563268052)", SupplierCategoryId: 2, BranchID: "BRC250118132203"},
		// {ID: "SPL250207144606", Name: "Bravo Supermarket", Phone: "(0321)855588", Address: "Jl. Yos Sudarso 78/92 Jombang", PIC: "-", SupplierCategoryId: 4, BranchID: "BRC250118132203"},
	}
	config.DB.Create(&supplier)
}
