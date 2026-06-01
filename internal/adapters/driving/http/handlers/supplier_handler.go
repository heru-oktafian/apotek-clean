package handlers

import (
	strings "strings"

	configs "apotek-clean/configs"
	helpers "apotek-clean/helpers"
	models "apotek-clean/models"
	services "apotek-clean/services"
	fiber "github.com/gofiber/fiber/v2"
)

type SupplierHandler struct{}

func NewSupplierHandler() *SupplierHandler {
	return &SupplierHandler{}
}

func (h *SupplierHandler) CreateSupplier(c *fiber.Ctx) error {
	return helpers.CreateResource(c, configs.DB, &models.Supplier{}, "SPL")
}

func (h *SupplierHandler) UpdateSupplier(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.UpdateResource(c, configs.DB, &models.Supplier{}, id)
}

func (h *SupplierHandler) DeleteSupplier(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.DeleteResource(c, configs.DB, &models.Supplier{}, id)
}

func (h *SupplierHandler) GetSupplierByID(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.GetResource(c, configs.DB, &models.Supplier{}, id)
}

func (h *SupplierHandler) GetAllSupplier(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	var supplier []models.SupplierDetail
	query := configs.DB.Table("suppliers s").
		Select("s.id, s.name, s.phone, s.address, s.pic, s.supplier_category_id, sc.name AS supplier_category").
		Joins("LEFT JOIN supplier_categories sc ON sc.id = s.supplier_category_id").
		Where("s.branch_id = ?", branchID)
	_, search, total, page, totalPages, limit, err := helpers.Paginate(c, query, &supplier, []string{"s.name ILIKE ?", "s.address ILIKE ?", "sc.name ILIKE ?"})
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get Supplier Category failed", err.Error())
	}
	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Suppliers retrieved successfully", search, int(total), page, int(totalPages), int(limit), supplier)
}

func (h *SupplierHandler) CmbSupplier(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	search := strings.ToLower(c.Query("search"))
	var cmbSuppliers []models.CmbSupplierModel
	query := configs.DB.Table("suppliers").
		Select("id AS supplier_id, name AS supplier_name").
		Where("branch_id = ?", branchID)
	if search != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+search+"%")
	}
	query = query.Order("name ASC")
	if err := query.Find(&cmbSuppliers).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to get data", err.Error())
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "Data berhasil ditemukan", cmbSuppliers)
}
