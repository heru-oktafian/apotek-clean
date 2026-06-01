package seeders

import (
	config "github.com/heru-oktafian/fiber-apotek/configs"
	models "github.com/heru-oktafian/fiber-apotek/models"
)

// UnitConversionSeed inisialisasi unit conversion
func UnitConversionSeed() {
	unitConversion := []models.UnitConversion{
		// {ID: "UNC250118132203", ProductId: "PRD250118132203", InitId: "UNT250118123204", FinalId: "UNT250118123203", ValueConv: 10, BranchID: "BRC250118132203"},
		// {ID: "UNC250118132204", ProductId: "PRD250118132203", InitId: "UNT250118123205", FinalId: "UNT250118123203", ValueConv: 100, BranchID: "BRC250118132203"},
		// {ID: "UNC250118132205", ProductId: "PRD250118132203", InitId: "UNT250118123205", FinalId: "UNT250118123204", ValueConv: 10, BranchID: "BRC250118132203"},
	}
	config.DB.Create(&unitConversion)
}
