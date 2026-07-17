package seeders

import (
	config "apotek-clean/configs"
	models "apotek-clean/internal/core/entities"
)

func MemberCategorySeed() {
	memberCategory := []models.MemberCategory{
		// {Name: "Reguler", BranchID: "BRC250118132203"},
		// {Name: "Silver", BranchID: "BRC250118132203"},
		// {Name: "Gold", BranchID: "BRC250118132203"},
	}
	config.DB.Create(&memberCategory)
}
