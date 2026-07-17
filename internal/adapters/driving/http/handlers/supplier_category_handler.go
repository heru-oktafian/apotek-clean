package handlers

import (
	configs "apotek-clean/configs"
	helpers "apotek-clean/helpers"
	models "apotek-clean/internal/core/entities"
	services "apotek-clean/services"
	fiber "github.com/gofiber/fiber/v2"
)

type SupplierCategoryHandler struct{}

func NewSupplierCategoryHandler() *SupplierCategoryHandler {
	return &SupplierCategoryHandler{}
}

func (h *SupplierCategoryHandler) CreateSupplierCategory(c *fiber.Ctx) error {
	return helpers.CreateResourceInc(c, configs.DB, &models.SupplierCategory{})
}

func (h *SupplierCategoryHandler) UpdateSupplierCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.UpdateResource(c, configs.DB, &models.SupplierCategory{}, id)
}

func (h *SupplierCategoryHandler) DeleteSupplierCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.DeleteResource(c, configs.DB, &models.SupplierCategory{}, id)
}

func (h *SupplierCategoryHandler) GetSupplierCategoryByID(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.GetResource(c, configs.DB, &models.SupplierCategory{}, id)
}

func (h *SupplierCategoryHandler) GetAllSupplierCategory(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	var supplierCategory []models.SupplierCategory
	query := configs.DB.Table("supplier_categories sc").
		Select("sc.id, sc.name, sc.branch_id").
		Where("sc.branch_id = ?", branchID).
		Order("sc.name ASC")
	_, search, total, page, totalPages, limit, err := helpers.Paginate(c, query, &supplierCategory, []string{"sc.name ILIKE ?"})
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get Supplier Category failed", err.Error())
	}
	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Supplier Category retrieved successfully", search, int(total), page, int(totalPages), int(limit), supplierCategory)
}

func (h *SupplierCategoryHandler) CmbSupplierCategory(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	var cmbSupplierCategories []models.SupplierCategoryCombo
	if err := configs.DB.Table("supplier_categories").
		Select("id AS supplier_category_id, name AS supplier_category_name").
		Where("branch_id = ?", branchID).
		Order("name ASC").
		Find(&cmbSupplierCategories).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to get data", "Failed to get data")
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "Data berhasil ditemukan", cmbSupplierCategories)
}
