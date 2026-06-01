package controllers

import (
	fiber "github.com/gofiber/fiber/v2"
	configs "github.com/heru-oktafian/fiber-apotek/configs"
	helpers "github.com/heru-oktafian/fiber-apotek/helpers"
	models "github.com/heru-oktafian/fiber-apotek/models"
	services "github.com/heru-oktafian/fiber-apotek/services"
)

// CreateSupplierCategory buat Supplier category
func CreateSupplierCategory(c *fiber.Ctx) error {
	// Creating new SupplierCategory using helpers
	return helpers.CreateResourceInc(c, configs.DB, &models.SupplierCategory{})
}

// UpdateSupplierCategory update SupplierCategory
func UpdateSupplierCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	// Updating SupplierCategory using helpers
	return helpers.UpdateResource(c, configs.DB, &models.SupplierCategory{}, id)
}

// DeleteSupplierCategory hapus SupplierCategory
func DeleteSupplierCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	// Deleting SupplierCategory using helpers
	return helpers.DeleteResource(c, configs.DB, &models.SupplierCategory{}, id)
}

// GetSupplierCategoryByID tampilkan SupplierCategory berdasarkan id
func GetSupplierCategoryByID(c *fiber.Ctx) error {
	id := c.Params("id")
	// Getting SupplierCategory using helpers
	return helpers.GetResource(c, configs.DB, &models.SupplierCategory{}, id)
}

// GetSupplierCategory tampilkan SupplierCategory
func GetAllSupplierCategory(c *fiber.Ctx) error {
	// Get branch id
	branch_id, _ := services.GetBranchID(c)

	var SupplierCategory []models.SupplierCategory

	// Query dasar
	query := configs.DB.Table("supplier_categories sc").
		Select("sc.id, sc.name, sc.branch_id").
		Where("sc.branch_id = ?", branch_id).
		Order("sc.name ASC")

	_, search, total, page, totalPages, limit, err := helpers.Paginate(c, query, &SupplierCategory, []string{"sc.name ILIKE ?"})
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get Supplier Category failed", err.Error())
	}

	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Supplier Category retrieved successfully", search, int(total), page, int(totalPages), int(limit), SupplierCategory)
}

// CmbSupplierCategory mendapatkan semua kategori supplier
func CmbSupplierCategory(c *fiber.Ctx) error {
	// Get branch id
	branch_id, _ := services.GetBranchID(c)

	var cmbSupplierCategories []models.SupplierCategoryCombo

	// Query untuk mendapatkan semua kategori supplier
	if err := configs.DB.Table("supplier_categories").
		Select("id AS supplier_category_id, name AS supplier_category_name").
		Where("branch_id = ?", branch_id).
		Order("name ASC").
		Find(&cmbSupplierCategories).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to get data", "Failed to get data")
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Data berhasil ditemukan", cmbSupplierCategories)
}
