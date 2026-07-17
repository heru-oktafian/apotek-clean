package handlers

import (
	strings "strings"

	configs "apotek-clean/configs"
	helpers "apotek-clean/helpers"
	models "apotek-clean/internal/core/entities"
	services "apotek-clean/services"
	fiber "github.com/gofiber/fiber/v2"
)

type MemberHandler struct{}

func NewMemberHandler() *MemberHandler {
	return &MemberHandler{}
}

func (h *MemberHandler) CreateMember(c *fiber.Ctx) error {
	return helpers.CreateResource(c, configs.DB, &models.Member{}, "MBR")
}

func (h *MemberHandler) UpdateMember(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.UpdateResource(c, configs.DB, &models.Member{}, id)
}

func (h *MemberHandler) DeleteMember(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.DeleteResource(c, configs.DB, &models.Member{}, id)
}

func (h *MemberHandler) GetMember(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.GetResource(c, configs.DB, &models.Member{}, id)
}

func (h *MemberHandler) GetAllMember(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	var member []models.MemberDetail
	query := configs.DB.Table("members m").
		Select("m.id, m.name, m.phone, m.address, m.member_category_id, mc.name AS member_category, m.points").
		Joins("LEFT JOIN member_categories mc ON mc.id = m.member_category_id").
		Where("m.branch_id = ?", branchID)
	_, search, total, page, totalPages, limit, err := helpers.Paginate(c, query, &member, []string{"m.name ILIKE ?", "m.phone ILIKE ?", "m.address ILIKE ?", "mc.name ILIKE ?"})
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get Units failed", err.Error())
	}
	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Data berhasil ditemukan", search, int(total), page, int(totalPages), int(limit), member)
}

func (h *MemberHandler) CmbMember(c *fiber.Ctx) error {
	search := strings.TrimSpace(c.Query("search"))
	branchID, _ := services.GetBranchID(c)
	var members []models.ComboboxMembers
	query := configs.DB.Table("members").Select("id AS member_id, name AS member_name").Where("branch_id = ?", branchID)
	if search != "" {
		search = strings.ToLower(search)
		query = query.Where("LOWER(members.name) ILIKE ?", "%"+search+"%")
	}
	if err := query.Find(&members).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to get data", "Failed to get data")
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "Data berhasil ditemukan", members)
}
