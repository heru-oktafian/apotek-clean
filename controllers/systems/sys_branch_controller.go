package controllers

import (
	fiber "github.com/gofiber/fiber/v2"
	config "github.com/heru-oktafian/fiber-apotek/configs"
	helpers "github.com/heru-oktafian/fiber-apotek/helpers"
	models "github.com/heru-oktafian/fiber-apotek/models"
	services "github.com/heru-oktafian/fiber-apotek/services"
)

// CreateBranch adalah fungsi untuk membuat cabang baru
func CreateBranch(c *fiber.Ctx) error {
	// Membuat cabang baru menggunakan helpers
	return helpers.CreateResource(c, config.DB, &models.Branch{}, "BRC")
}

// UpdateBranch adalah fungsi untuk memperbarui cabang
func UpdateBranch(c *fiber.Ctx) error {
	id := c.Params("id")
	// Memperbarui cabang menggunakan helpers
	return helpers.UpdateResource(c, config.DB, &models.Branch{}, id)
}

// DeleteBranch adalah fungsi untuk menghapus cabang
func DeleteBranch(c *fiber.Ctx) error {
	id := c.Params("id")
	// Menghapus cabang menggunakan helpers
	return helpers.DeleteResource(c, config.DB, &models.Branch{}, id)
}

// GetBranch adalah fungsi untuk mendapatkan cabang
func GetBranch(c *fiber.Ctx) error {
	id := c.Params("id")
	// Mengambil cabang menggunakan helpers
	return helpers.GetResource(c, config.DB, &models.Branch{}, id)
}

// GetAllBranch adalah fungsi untuk mendapatkan semua cabang
func GetAllBranch(c *fiber.Ctx) error {
	var branches []models.Branch
	// Mengambil semua cabang menggunakan helpers
	return helpers.GetAllBranches(c, config.DB, &branches)
}

// GetUserBranch menangani penampilan userbranch
func GetBranchByUserId(c *fiber.Ctx) error {
	// dapatkan user id
	userID, _ := services.GetUserID(c)

	// Menampilkan semua userbranch
	var userBranchDetails []models.UserBranchDetail

	// Melakukan LEFT OUTER JOIN menggunakan GORM
	if err := config.DB.
		Table("user_branches").
		Select("user_branches.user_id, users.name AS user_name, user_branches.branch_id, branches.branch_name, branches.sia_name, branches.sipa_name, branches.phone").
		Joins("LEFT JOIN users ON users.id = user_branches.user_id").
		Joins("LEFT JOIN branches ON branches.id = user_branches.branch_id").
		Where("branches.branch_status = 'active' AND user_branches.user_id = ?", userID).
		Scan(&userBranchDetails).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get userbranches failed", "Failed to fetch user branches with details")
	}

	// Mengembalikan response data userbranch
	return helpers.JSONResponse(c, fiber.StatusOK, "User Branch found", userBranchDetails)
}
