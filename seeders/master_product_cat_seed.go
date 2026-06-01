package seeders

import (
	config "github.com/heru-oktafian/fiber-apotek/configs"
	models "github.com/heru-oktafian/fiber-apotek/models"
)

func ProductCategorySeed() {
	productCategory := []models.ProductCategory{
		// {Name: "Obat", BranchID: "BRC250118132203"},
		// {Name: "Vitamin", BranchID: "BRC250118132203"},
		// {Name: "Suplemen", BranchID: "BRC250118132203"},
		// {Name: "Susu", BranchID: "BRC250118132203"},
	}
	config.DB.Create(&productCategory)
}
