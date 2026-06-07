package handlers

import (
	configs "apotek-clean/configs"
	helpers "apotek-clean/helpers"
	models "apotek-clean/models"
	services "apotek-clean/services"
	fiber "github.com/gofiber/fiber/v2"
)

type UserBranchHandler struct{}

func NewUserBranchHandler() *UserBranchHandler {
	return &UserBranchHandler{}
}

func (h *UserBranchHandler) CreateUserBranch(c *fiber.Ctx) error {
	var userBranch models.UserBranch
	if err := c.BodyParser(&userBranch); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid input", err)
	}
	if err := configs.DB.Create(&userBranch).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to create user", err)
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "UserBranch created successfully", userBranch)
}

func (h *UserBranchHandler) GetUserBranch(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	branchID, _ := services.GetBranchID(c)
	var userBranchDetails []models.UserBranchDetail
	if err := configs.DB.Table("user_branches").
		Select("user_branches.user_id, users.name AS user_name, user_branches.branch_id, branches.branch_name, branches.address, branches.phone").
		Joins("LEFT JOIN users ON users.id = user_branches.user_id").
		Joins("LEFT JOIN branches ON branches.id = user_branches.branch_id").
		Where("branches.branch_status = 'active' AND user_branches.branch_id = ? AND user_branches.user_id = ?", branchID, userID).
		Scan(&userBranchDetails).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get userbranches failed", "Failed to fetch user branches with details")
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "UserBranch found", userBranchDetails)
}

func (h *UserBranchHandler) UpdateUserBranch(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	branchID, _ := services.GetBranchID(c)
	var userBranch models.UserBranch
	if err := configs.DB.Where("user_id = ? AND branch_id = ?", userID, branchID).First(&userBranch).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "UserBranch not found", err)
	}
	if err := c.BodyParser(&userBranch); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid input", err)
	}
	if err := configs.DB.Model(&userBranch).Where("user_id = ? AND branch_id = ?", userID, branchID).Updates(userBranch).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to update userbranch", err)
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "UserBranch updated successfully", userBranch)
}

func (h *UserBranchHandler) DeleteUserBranch(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	branchID := c.Params("branch_id")
	var userBranch models.UserBranch
	if err := configs.DB.Where("user_id = ? AND branch_id = ?", userID, branchID).First(&userBranch).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "userbranch not found", err)
	}
	if err := configs.DB.Unscoped().Where("user_id = ? AND branch_id = ?", userID, branchID).Delete(&userBranch).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to delete userbranch permanently", err)
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "UserBranch deleted successfully", userBranch)
}

func (h *UserBranchHandler) GetAllUserBranch(c *fiber.Ctx) error {
	var userBranchDetails []models.UserBranchDetail
	if err := configs.DB.Table("user_branches usrb").
		Select("usrb.user_id, usr.name AS user_name, usrb.branch_id, brc.branch_name AS branch_name, brc.sia_name, brc.sipa_name, brc.phone").
		Joins("LEFT JOIN users usr ON usr.id = usrb.user_id").
		Joins("LEFT JOIN branches brc ON brc.id = usrb.branch_id").
		Scan(&userBranchDetails).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get user branches failed", "Failed to fetch user branches with details")
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "UserBranches retrieved successfully", userBranchDetails)
}

func (h *UserBranchHandler) GetUserDetails(c *fiber.Ctx) error {
	userID := c.Params("id")
	var user models.User
	if err := configs.DB.First(&user, "id = ?", userID).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Pengguna tidak ditemukan", err)
	}
	var userBranches []models.UserBranch
	if err := configs.DB.Where("user_id = ?", userID).Find(&userBranches).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mendapatkan cabang", err)
	}
	var branchIDs []string
	for _, ub := range userBranches {
		branchIDs = append(branchIDs, ub.BranchID)
	}
	var branches []models.Branch
	if len(branchIDs) > 0 {
		if err := configs.DB.Where("id IN ?", branchIDs).Find(&branches).Error; err != nil {
			return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal memuat detail cabang", err)
		}
	}
	type BranchResponse struct {
		BranchID   string `json:"branch_id"`
		BranchName string `json:"branch_name"`
		Address    string `json:"address"`
		Phone      string `json:"phone"`
	}
	var branchResponses []BranchResponse
	for _, b := range branches {
		branchResponses = append(branchResponses, BranchResponse{BranchID: b.ID, BranchName: b.BranchName, Address: b.Address, Phone: b.Phone})
	}
	type GetUserResponse struct {
		User           models.User      `json:"user"`
		DetailBranches []BranchResponse `json:"detail_branches"`
	}
	response := GetUserResponse{User: user, DetailBranches: branchResponses}
	return helpers.JSONResponse(c, fiber.StatusOK, "Data berhasil ditemukan", response)
}
