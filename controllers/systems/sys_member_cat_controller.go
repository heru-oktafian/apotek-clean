package controllers

import (
	strings "strings"

	fiber "github.com/gofiber/fiber/v2"
	configs "github.com/heru-oktafian/fiber-apotek/configs"
	helpers "github.com/heru-oktafian/fiber-apotek/helpers"
	models "github.com/heru-oktafian/fiber-apotek/models"
	services "github.com/heru-oktafian/fiber-apotek/services"
)

// CreateMemberCategory buat member category
func CreateMemberCategory(c *fiber.Ctx) error {
	// Creating new MemberCategory using helpers
	return helpers.CreateResourceInc(c, configs.DB, &models.MemberCategory{})
}

// UpdateMemberCategory update MemberCategory
func UpdateMemberCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	// Updating MemberCategory using helpers
	return helpers.UpdateResource(c, configs.DB, &models.MemberCategory{}, id)
}

// DeleteMemberCategory hapus MemberCategory
func DeleteMemberCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	// Deleting MemberCategory using helpers
	return helpers.DeleteResource(c, configs.DB, &models.MemberCategory{}, id)
}

// GetMemberCategory tampilkan MemberCategory berdasarkan id
func GetMemberCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	// Getting MemberCategory using helpers
	return helpers.GetResource(c, configs.DB, &models.MemberCategory{}, id)
}

// GetAllMemberCategories tampilkan semua MemberCategory
func GetAllMemberCategory(c *fiber.Ctx) error {
	// Get branch id
	branch_id, _ := services.GetBranchID(c)

	var MemberCategory []models.MemberCategory

	// Query dasar
	query := configs.DB.Table("member_categories mc").Select("mc.id, mc.name, mc.points_conversion_rate, mc.branch_id").Where("mc.branch_id = ?", branch_id)

	_, search, total, page, totalPages, limit, err := helpers.Paginate(c, query, &MemberCategory, []string{"mc.name ILIKE ?"})
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get Units failed", err.Error())
	}

	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Member Categories retrieved successfully", search, int(total), page, int(totalPages), int(limit), MemberCategory)

}

// CmbMemberCategory mendapatkan semua kategori member
func CmbMemberCategory(c *fiber.Ctx) error {
	// Parsing query parameters
	search := strings.TrimSpace(c.Query("search"))

	// Get branch id
	branch_id, _ := services.GetBranchID(c)

	var categories []models.ComboMemberCategory

	// Query untuk mendapatkan semua kategori member
	query := configs.DB.Table("member_categories").
		Select("id AS member_category_id, name AS member_category_name").
		Where("branch_id = ?", branch_id)

	// Jika ada search key, tambahkan filter WHERE
	if search != "" {
		search = strings.ToLower(search) // Konversi search ke lowercase
		query = query.Where("LOWER(member_categories.name) ILIKE ?", "%"+search+"%")
	}

	if err := query.Find(&categories).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to get data", "Failed to get data")
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Data berhasil ditemukan", categories)
}
