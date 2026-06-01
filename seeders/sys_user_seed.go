package seeders

import (
	config "github.com/heru-oktafian/fiber-apotek/configs"
	models "github.com/heru-oktafian/fiber-apotek/models"
)

func UserSeed() {
	users := []models.User{
		// {ID: "USR250118132201", Username: "vita_fauzi", Password: "Sigala1102", Name: "Vita Fauzi. M", UserRole: "superadmin", UserStatus: "active"},
		// {ID: "USR250118132202", Username: "fanny", Password: "Izahfanny17", Name: "Fanny", UserRole: "operator", UserStatus: "active"},
		// {ID: "USR250118132203", Username: "zia", Password: "zia123", Name: "Zia", UserRole: "operator", UserStatus: "active"},
	}

	// Hash password untuk setiap user
	for _, user := range users {
		if err := user.HashPassword(); err != nil {
			continue
		}
		config.DB.Create(&user)
	}
}
