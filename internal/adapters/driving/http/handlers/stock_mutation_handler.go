package handlers

import (
	"strconv"

	"apotek-clean/configs"
	"apotek-clean/helpers"
	"apotek-clean/services"

	"github.com/gofiber/fiber/v2"
)

type StockMutationHandler struct{}

func NewStockMutationHandler() *StockMutationHandler {
	return &StockMutationHandler{}
}

type CreateStockMutationRequest struct {
	ProductID string `json:"product_id"`
	FromLoc   string `json:"from_loc"`
	ToLoc     string `json:"to_loc"`
	Qty       int    `json:"qty"`
	Note      string `json:"note"`
}

func (h *StockMutationHandler) CreateStockMutation(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	userID, _ := services.GetUserID(c)

	var req CreateStockMutationRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if req.ProductID == "" {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "product_id is required", nil)
	}
	if req.Qty <= 0 {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "qty must be greater than 0", nil)
	}
	if req.FromLoc == req.ToLoc {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "from_loc and to_loc cannot be the same", nil)
	}
	validLoc := map[string]bool{"warehouse": true, "showcase": true}
	if !validLoc[req.FromLoc] || !validLoc[req.ToLoc] {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "invalid location: use 'warehouse' or 'showcase'", nil)
	}

	err := services.CreateStockMutation(configs.DB, req.ProductID, req.FromLoc, req.ToLoc, req.Qty, req.Note, userID, branchID)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return helpers.JSONResponse(c, fiber.StatusCreated, "Stock mutation created successfully", nil)
}

func (h *StockMutationHandler) GetStockMutations(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	productID := c.Query("product_id")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	mutations, total, err := services.GetStockMutations(configs.DB, branchID, page, limit, productID)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to get stock mutations", err.Error())
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Stock mutations retrieved successfully", "", int(total), page, totalPages, limit, mutations)
}
