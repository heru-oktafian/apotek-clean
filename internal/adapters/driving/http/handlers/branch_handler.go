package handlers

import (
	config "apotek-clean/configs"
	helpers "apotek-clean/helpers"
	models "apotek-clean/models"
	services "apotek-clean/services"
	fiber "github.com/gofiber/fiber/v2"
)

type BranchHandler struct{}

func NewBranchHandler() *BranchHandler {
	return &BranchHandler{}
}

func (h *BranchHandler) CreateBranch(c *fiber.Ctx) error {
	return helpers.CreateResource(c, config.DB, &models.Branch{}, "BRC")
}

func (h *BranchHandler) UpdateBranch(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.UpdateResource(c, config.DB, &models.Branch{}, id)
}

func (h *BranchHandler) DeleteBranch(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.DeleteResource(c, config.DB, &models.Branch{}, id)
}

func (h *BranchHandler) GetBranch(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.GetResource(c, config.DB, &models.Branch{}, id)
}

func (h *BranchHandler) GetAllBranch(c *fiber.Ctx) error {
	var branches []models.Branch
	return helpers.GetAllBranches(c, config.DB, &branches)
}

func (h *BranchHandler) GetBranchByUserId(c *fiber.Ctx) error {
	userID, _ := services.GetUserID(c)
	userBranchDetails, err := services.GetActiveUserBranchDetails(config.DB, userID)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get userbranches failed", "Failed to fetch user branches with details")
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "User Branch found", userBranchDetails)
}
