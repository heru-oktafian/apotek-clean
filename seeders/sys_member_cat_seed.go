package seeders

import (
	config "github.com/heru-oktafian/fiber-apotek/configs"
	models "github.com/heru-oktafian/fiber-apotek/models"
)

func MemberCategorySeed() {
	memberCategory := []models.MemberCategory{
		// {Name: "Reguler", BranchID: "BRC250118132203"},
		// {Name: "Silver", BranchID: "BRC250118132203"},
		// {Name: "Gold", BranchID: "BRC250118132203"},
	}
	config.DB.Create(&memberCategory)
}
