package handlers

import (
	strings "strings"

	config "apotek-clean/configs"
	helpers "apotek-clean/helpers"
	models "apotek-clean/models"
	services "apotek-clean/services"
	fiber "github.com/gofiber/fiber/v2"
)

type ProductCategoryHandler struct{}

func NewProductCategoryHandler() *ProductCategoryHandler {
	return &ProductCategoryHandler{}
}

func (h *ProductCategoryHandler) CreateProductCategory(c *fiber.Ctx) error {
	return helpers.CreateResourceInc(c, config.DB, &models.ProductCategory{})
}

func (h *ProductCategoryHandler) UpdateProductCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.UpdateResource(c, config.DB, &models.ProductCategory{}, id)
}

func (h *ProductCategoryHandler) DeleteProductCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.DeleteResource(c, config.DB, &models.ProductCategory{}, id)
}

func (h *ProductCategoryHandler) GetProductCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.GetResource(c, config.DB, &models.ProductCategory{}, id)
}

func (h *ProductCategoryHandler) CmbProductCategory(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	search := strings.TrimSpace(c.Query("search"))
	var categories []models.ComboProductCategory
	query := config.DB.Table("product_categories").
		Select("product_categories.id as product_category_id, product_categories.name as product_category_name").
		Where("branch_id = ?", branchID)
	if search != "" {
		search = strings.ToLower(search)
		query = query.Where("LOWER(product_categories.name) LIKE ?", "%"+search+"%")
	}
	query = query.Order("product_categories.name ASC")
	if err := query.Find(&categories).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to get data", "Failed to get data")
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "Data berhasil ditemukan", categories)
}

func (h *ProductCategoryHandler) GetAllProductCategory(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	var productCategory []models.ComboProductCategory
	query := config.DB.Table("product_categories pc").Select("pc.id AS product_category_id, pc.name AS product_category_name").Where("pc.branch_id = ?", branchID)
	_, search, total, page, totalPages, limit, err := helpers.Paginate(c, query, &productCategory, []string{"pc.name ILIKE ?"})
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get Product Categories failed", err.Error())
	}
	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Product Categories retrieved successfully", search, total, page, totalPages, limit, productCategory)
}
