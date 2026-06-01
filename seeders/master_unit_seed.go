package seeders

import (
	config "github.com/heru-oktafian/fiber-apotek/configs"
	models "github.com/heru-oktafian/fiber-apotek/models"
)

// UnitSeed inisialisasi unit
func UnitSeed() {
	unit := []models.Unit{
		// {ID: "UNT250118123203", Name: "Pcs", BranchID: "BRC250118132203"},
		// {ID: "UNT250118123204", Name: "Strip", BranchID: "BRC250118132203"},
		// {ID: "UNT250118123205", Name: "Box", BranchID: "BRC250118132203"},
		// {ID: "UNT250118123221", Name: "Tablet", BranchID: "BRC250118132203"},
		// {ID: "UNT250118123222", Name: "Kapsul", BranchID: "BRC250118132203"},
		// {ID: "UNT250118123210", Name: "Kaplet", BranchID: "BRC250118132203"},
		// {ID: "UNT250118123206", Name: "Tube", BranchID: "BRC250118132203"},
		// {ID: "UNT250118123207", Name: "Fls", BranchID: "BRC250118132203"},
		// {ID: "UNT250118123208", Name: "Sachet", BranchID: "BRC250118132203"},
		// {ID: "UNT250118123209", Name: "Botol", BranchID: "BRC250118132203"},
		// {ID: "UNT250118123211", Name: "Batang", BranchID: "BRC250118132203"},
		// {ID: "UNT250118123212", Name: "Pac", BranchID: "BRC250118132203"},
		// {ID: "UNT250118123213", Name: "Pot", BranchID: "BRC250118132203"},
		// {ID: "UNT250118123214", Name: "Dus", BranchID: "BRC250118132203"},
		// {ID: "UNT250118123215", Name: "Kaleng", BranchID: "BRC250118132203"},
		// {ID: "UNT250118123216", Name: "Ampul", BranchID: "BRC250118132203"},
		// {ID: "UNT250118123217", Name: "Renteng", BranchID: "BRC250118132203"},
		// {ID: "UNT250118123218", Name: "Bungkus", BranchID: "BRC250118132203"},
		// {ID: "UNT250118123219", Name: "Rol", BranchID: "BRC250118132203"},
		// {ID: "UNT250118123223", Name: "Kg", BranchID: "BRC250118132203"},
		// {ID: "UNT250118123224", Name: "Supp", BranchID: "BRC250118132203"},
	}
	config.DB.Create(&unit)
}
