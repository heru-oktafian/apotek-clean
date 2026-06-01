package seeders

import (
	config "github.com/heru-oktafian/fiber-apotek/configs"
	models "github.com/heru-oktafian/fiber-apotek/models"
)

func UserBranchSeed() {
	userBranch := []models.UserBranch{
		// {UserID: "USR250118132201", BranchID: "BRC250118132203"},
		// {UserID: "USR250118132202", BranchID: "BRC250118132203"},
		// {UserID: "USR250118132203", BranchID: "BRC250118132203"},
	}
	config.DB.Create(&userBranch)
}
