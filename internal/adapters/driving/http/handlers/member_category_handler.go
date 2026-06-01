package handlers

import (
	strings "strings"

	configs "apotek-clean/configs"
	helpers "apotek-clean/helpers"
	models "apotek-clean/models"
	services "apotek-clean/services"
	fiber "github.com/gofiber/fiber/v2"
)

type MemberCategoryHandler struct{}

func NewMemberCategoryHandler() *MemberCategoryHandler {
	return &MemberCategoryHandler{}
}

func (h *MemberCategoryHandler) CreateMemberCategory(c *fiber.Ctx) error {
	return helpers.CreateResourceInc(c, configs.DB, &models.MemberCategory{})
}

func (h *MemberCategoryHandler) UpdateMemberCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.UpdateResource(c, configs.DB, &models.MemberCategory{}, id)
}

func (h *MemberCategoryHandler) DeleteMemberCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.DeleteResource(c, configs.DB, &models.MemberCategory{}, id)
}

func (h *MemberCategoryHandler) GetMemberCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.GetResource(c, configs.DB, &models.MemberCategory{}, id)
}

func (h *MemberCategoryHandler) GetAllMemberCategory(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	var memberCategory []models.MemberCategory
	query := configs.DB.Table("member_categories mc").Select("mc.id, mc.name, mc.points_conversion_rate, mc.branch_id").Where("mc.branch_id = ?", branchID)
	_, search, total, page, totalPages, limit, err := helpers.Paginate(c, query, &memberCategory, []string{"mc.name ILIKE ?"})
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get Units failed", err.Error())
	}
	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Member Categories retrieved successfully", search, int(total), page, int(totalPages), int(limit), memberCategory)
}

func (h *MemberCategoryHandler) CmbMemberCategory(c *fiber.Ctx) error {
	search := strings.TrimSpace(c.Query("search"))
	branchID, _ := services.GetBranchID(c)
	var categories []models.ComboMemberCategory
	query := configs.DB.Table("member_categories").Select("id AS member_category_id, name AS member_category_name").Where("branch_id = ?", branchID)
	if search != "" {
		search = strings.ToLower(search)
		query = query.Where("LOWER(member_categories.name) ILIKE ?", "%"+search+"%")
	}
	if err := query.Find(&categories).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to get data", "Failed to get data")
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "Data berhasil ditemukan", categories)
}
