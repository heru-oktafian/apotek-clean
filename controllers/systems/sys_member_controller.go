package controllers

import (
	strings "strings"

	fiber "github.com/gofiber/fiber/v2"
	configs "github.com/heru-oktafian/fiber-apotek/configs"
	helpers "github.com/heru-oktafian/fiber-apotek/helpers"
	models "github.com/heru-oktafian/fiber-apotek/models"
	services "github.com/heru-oktafian/fiber-apotek/services"
)

// CreateMember buat Member
func CreateMember(c *fiber.Ctx) error {
	// Creating new Member using helpers
	return helpers.CreateResource(c, configs.DB, &models.Member{}, "MBR")
}

// UpdateMember update Member
func UpdateMember(c *fiber.Ctx) error {
	id := c.Params("id")
	// Updating Member using helpers
	return helpers.UpdateResource(c, configs.DB, &models.Member{}, id)
}

// DeleteMember hapus Member
func DeleteMember(c *fiber.Ctx) error {
	id := c.Params("id")
	// Deleting Member using helpers
	return helpers.DeleteResource(c, configs.DB, &models.Member{}, id)
}

// GetMember tampilkan Member berdasarkan id
func GetMember(c *fiber.Ctx) error {
	id := c.Params("id")
	// Getting Member using helpers
	return helpers.GetResource(c, configs.DB, &models.Member{}, id)
}

// GetAllMember tampilkan semua Member
func GetAllMember(c *fiber.Ctx) error {
	// Ambil ID cabang
	branch_id, _ := services.GetBranchID(c)

	// Inisiasi variabel Member untuk memuat model member
	var Member []models.MemberDetail

	// Query dasar
	query := configs.DB.Table("members m").
		Select("m.id, m.name, m.phone, m.address, m.member_category_id, mc.name AS member_category, m.points").
		Joins("LEFT JOIN member_categories mc ON mc.id = m.member_category_id").
		Where("m.branch_id = ?", branch_id)

	_, search, total, page, totalPages, limit, err := helpers.Paginate(c, query, &Member, []string{"m.name ILIKE ?", "m.phone ILIKE ?", "m.address ILIKE ?", "mc.name ILIKE ?"})
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get Units failed", err.Error())
	}

	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Data berhasil ditemukan", search, int(total), page, int(totalPages), int(limit), Member)
}

// CmbMember mendapatkan semua member
func CmbMember(c *fiber.Ctx) error {
	// Parsing query parameters
	search := strings.TrimSpace(c.Query("search"))

	// Get branch id
	branch_id, _ := services.GetBranchID(c)

	var members []models.ComboboxMembers

	// Query untuk mendapatkan semua member
	query := configs.DB.Table("members").
		Select("id AS member_id, name AS member_name").
		Where("branch_id = ?", branch_id)

	// Jika ada search key, tambahkan filter WHERE
	if search != "" {
		search = strings.ToLower(search) // Konversi search ke lowercase
		query = query.Where("LOWER(members.name) ILIKE ?", "%"+search+"%")
	}

	if err := query.Find(&members).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to get data", "Failed to get data")
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Data berhasil ditemukan", members)
}
