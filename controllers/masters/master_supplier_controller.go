package controllers

import (
	strings "strings"

	fiber "github.com/gofiber/fiber/v2"
	configs "github.com/heru-oktafian/fiber-apotek/configs"
	helpers "github.com/heru-oktafian/fiber-apotek/helpers"
	models "github.com/heru-oktafian/fiber-apotek/models"
	services "github.com/heru-oktafian/fiber-apotek/services"
)

// CreateSupplier buat Supplier
func CreateSupplier(c *fiber.Ctx) error {
	// Creating new Supplier using helpers
	return helpers.CreateResource(c, configs.DB, &models.Supplier{}, "SPL")
}

// UpdateSupplier update Supplier
func UpdateSupplier(c *fiber.Ctx) error {
	id := c.Params("id")
	// Updating Supplier using helpers
	return helpers.UpdateResource(c, configs.DB, &models.Supplier{}, id)
}

// DeleteSupplier hapus Supplier
func DeleteSupplier(c *fiber.Ctx) error {
	id := c.Params("id")
	// Deleting Supplier using helpers
	return helpers.DeleteResource(c, configs.DB, &models.Supplier{}, id)
}

// GetSupplierByID tampilkan Supplier berdasarkan id
func GetSupplierByID(c *fiber.Ctx) error {
	id := c.Params("id")
	// Getting Supplier using helpers
	return helpers.GetResource(c, configs.DB, &models.Supplier{}, id)
}

// GetAllSuppliers tampilkan semua Supplier
func GetAllSupplier(c *fiber.Ctx) error {
	// Get branch id
	branch_id, _ := services.GetBranchID(c)

	// Inisialisasi slice untuk menampung hasil query
	var Supplier []models.SupplierDetail

	// Query dasar
	query := configs.DB.Table("suppliers s").
		Select("s.id, s.name, s.phone, s.address, s.pic, s.supplier_category_id, sc.name AS supplier_category").
		Joins("LEFT JOIN supplier_categories sc ON sc.id = s.supplier_category_id").
		Where("s.branch_id = ?", branch_id)

	_, search, total, page, totalPages, limit, err := helpers.Paginate(c, query, &Supplier, []string{"s.name ILIKE ?", "s.address ILIKE ?", "sc.name ILIKE ?"})
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get Supplier Category failed", err.Error())
	}

	// Mengembalikan response data Supplier
	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Suppliers retrieved successfully", search, int(total), page, int(totalPages), int(limit), Supplier)

}

// CmbSupplier mendapatkan semua supplier untuk combo box
func CmbSupplier(c *fiber.Ctx) error {
	// Get branch id
	branch_id, _ := services.GetBranchID(c)

	// Ambil parameter search dari query URL
	search := strings.ToLower(c.Query("search"))

	var cmbSuppliers []models.CmbSupplierModel

	// Query untuk mendapatkan semua supplier
	query := configs.DB.Table("suppliers").
		Select("id AS supplier_id, name AS supplier_name").
		Where("branch_id = ?", branch_id)
	// Jika ada search key, tambahkan filter WHERE
	if search != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+search+"%")
	}

	// Urutkan hasil secara ascending berdasarkan name
	query = query.Order("name ASC")

	if err := query.Find(&cmbSuppliers).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to get data", err.Error())
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Data berhasil ditemukan", cmbSuppliers)
}
