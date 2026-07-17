package seeders

import (
	config "apotek-clean/configs"
	models "apotek-clean/internal/core/entities"
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
