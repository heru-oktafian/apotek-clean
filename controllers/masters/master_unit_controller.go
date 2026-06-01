package controllers

import (
	strings "strings"

	fiber "github.com/gofiber/fiber/v2"
	configs "github.com/heru-oktafian/fiber-apotek/configs"
	helpers "github.com/heru-oktafian/fiber-apotek/helpers"
	models "github.com/heru-oktafian/fiber-apotek/models"
	services "github.com/heru-oktafian/fiber-apotek/services"
)

// CreateUnit buat unit
func CreateUnit(c *fiber.Ctx) error {
	// Membuat unit baru menggunakan helpers
	return helpers.CreateResource(c, configs.DB, &models.Unit{}, "UNT")
}

// UpdateUnit update unit
func UpdateUnit(c *fiber.Ctx) error {
	id := c.Params("id")
	// Memperbarui unit menggunakan helpers
	return helpers.UpdateResource(c, configs.DB, &models.Unit{}, id)
}

// DeleteUnit hapus unit
func DeleteUnit(c *fiber.Ctx) error {
	id := c.Params("id")
	// Menghapus unit menggunakan helpers
	return helpers.DeleteResource(c, configs.DB, &models.Unit{}, id)
}

// GetUnit tampilkan unit berdasarkan id
func GetUnit(c *fiber.Ctx) error {
	id := c.Params("id")
	// Mengambil unit menggunakan helpers
	return helpers.GetResource(c, configs.DB, &models.Unit{}, id)
}

// GetAllUnit tampilkan semua unit
func GetAllUnit(c *fiber.Ctx) error {
	// Get branch id
	branch_id, _ := services.GetBranchID(c)

	var Unit []models.Unit

	// Query dasar
	query := configs.DB.Table("units un").Select("un.id, un.name, un.branch_id").Where("un.branch_id = ?", branch_id)

	_, search, total, page, totalPages, limit, err := helpers.Paginate(c, query, &Unit, []string{"un.name ILIKE ?"})
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get Units failed", err.Error())
	}

	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Units retrieved successfully", search, total, page, totalPages, limit, Unit)
}

// CmbUnit mendapatkan semua kategori unit
func CmbUnit(c *fiber.Ctx) error {
	// Get branch id
	branch_id, _ := services.GetBranchID(c)

	// Parsing parameter query untuk pencarian
	search := strings.TrimSpace(c.Query("search"))

	var cmbUnits []models.UnitCombo

	// Query dasar untuk mendapatkan semua kategori unit
	query := configs.DB.Table("units").
		Select("id as unit_id, name as unit_name").
		Where("branch_id = ?", branch_id)

	// Jika parameter pencarian disediakan, tambahkan filter
	if search != "" {
		search = strings.ToLower(search) // Konversi kata kunci pencarian ke huruf kecil
		query = query.Where("LOWER(name) LIKE ?", "%"+search+"%")
	}

	// Tambahkan urutan berdasarkan nama secara ascending
	query = query.Order("name ASC")

	// Eksekusi query
	if err := query.Find(&cmbUnits).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to get data", err.Error())
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Data berhasil ditemukan", cmbUnits)
}
