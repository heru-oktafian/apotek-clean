package seeders

import (
	config "github.com/heru-oktafian/fiber-apotek/configs"
	models "github.com/heru-oktafian/fiber-apotek/models"
)

func BranchSeed() {
	// t := time.Date(2026, time.December, 31, 0, 0, 0, 0, time.UTC)

	branch := []models.Branch{
		// {ID: "BRC250118132203", BranchName: "Ziida Farma", Address: "Dusun Mireng, RT./RW. 001/003, Desa Sumberagung, Kecamatan Megaluh, Kabupaten Jombang, Jawa Timur", Phone: "085335833636", Email: "vita.alfarizqi@gmail.com", SiaId: "08012400152290001", SiaName: "Apotek Ziida Farma", PsaId: "3517011710880001", PsaName: "Vita Fauzi. M", Sipa: "446/065/415.35/2024", SipaName: "Vita Fauzi. M", JournalMethod: "automatic", TaxPercentage: 11, BranchStatus: "active", LicenseDate: t},
	}
	config.DB.Create(&branch)
}
