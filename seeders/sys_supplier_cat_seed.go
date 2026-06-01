package seeders

import (
	config "github.com/heru-oktafian/fiber-apotek/configs"
	models "github.com/heru-oktafian/fiber-apotek/models"
)

// SupplierCategorySeed function untuk menambahkan data supplier category ke database
func SupplierCategorySeed() {
	supplierCategory := []models.SupplierCategory{
		// {Name: "PBF", BranchID: "BRC250118132203"},
		// {Name: "Distributor", BranchID: "BRC250118132203"},
		// {Name: "Sub Distributor", BranchID: "BRC250118132203"},
		// {Name: "Toko / Retail", BranchID: "BRC250118132203"},
	}
	config.DB.Create(&supplierCategory)
}
