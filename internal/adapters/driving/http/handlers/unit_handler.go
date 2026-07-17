package handlers

import (
	strings "strings"

	configs "apotek-clean/configs"
	helpers "apotek-clean/helpers"
	models "apotek-clean/internal/core/entities"
	services "apotek-clean/services"
	fiber "github.com/gofiber/fiber/v2"
)

type UnitHandler struct{}

func NewUnitHandler() *UnitHandler {
	return &UnitHandler{}
}

func (h *UnitHandler) CreateUnit(c *fiber.Ctx) error {
	return helpers.CreateResource(c, configs.DB, &models.Unit{}, "UNT")
}

func (h *UnitHandler) UpdateUnit(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.UpdateResource(c, configs.DB, &models.Unit{}, id)
}

func (h *UnitHandler) DeleteUnit(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.DeleteResource(c, configs.DB, &models.Unit{}, id)
}

func (h *UnitHandler) GetUnit(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.GetResource(c, configs.DB, &models.Unit{}, id)
}

func (h *UnitHandler) GetAllUnit(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	var unit []models.Unit
	query := configs.DB.Table("units un").Select("un.id, un.name, un.branch_id").Where("un.branch_id = ?", branchID)
	_, search, total, page, totalPages, limit, err := helpers.Paginate(c, query, &unit, []string{"un.name ILIKE ?"})
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get Units failed", err.Error())
	}
	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Units retrieved successfully", search, total, page, totalPages, limit, unit)
}

func (h *UnitHandler) CmbUnit(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	search := strings.TrimSpace(c.Query("search"))
	var cmbUnits []models.UnitCombo
	query := configs.DB.Table("units").Select("id as unit_id, name as unit_name").Where("branch_id = ?", branchID)
	if search != "" {
		search = strings.ToLower(search)
		query = query.Where("LOWER(name) LIKE ?", "%"+search+"%")
	}
	query = query.Order("name ASC")
	if err := query.Find(&cmbUnits).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to get data", err.Error())
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "Data berhasil ditemukan", cmbUnits)
}
