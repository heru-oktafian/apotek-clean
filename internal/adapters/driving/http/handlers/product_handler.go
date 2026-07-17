package handlers

import (
	strings "strings"

	configs "apotek-clean/configs"
	helpers "apotek-clean/helpers"
	models "apotek-clean/internal/core/entities"
	services "apotek-clean/services"
	fiber "github.com/gofiber/fiber/v2"
)

type ProductHandler struct{}

func NewProductHandler() *ProductHandler {
	return &ProductHandler{}
}

func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	var product models.Product
	if err := c.BodyParser(&product); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid input", err)
	}
	product.ID = helpers.GenerateID("PRD")
	branchID, _ := services.GetBranchID(c)
	product.BranchID = branchID
	product.Stock = 0
	if strings.TrimSpace(product.SKU) == "" {
		product.SKU = product.ID
	}
	if err := configs.DB.Create(&product).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to create resource", err)
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "Resource created successfully", &product)
}

func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.UpdateResource(c, configs.DB, &models.Product{}, id)
}

func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.DeleteResource(c, configs.DB, &models.Product{}, id)
}

func (h *ProductHandler) GetProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	var allProduct []models.ProductDetail
	if err := configs.DB.
		Table("products pro").
		Select("pro.id,pro.sku,pro.name,pro.description, pro.ingredient, pro.dosage, pro.side_affection, pro.unit_id AS unit_id,pro.stock,pro.purchase_price,pro.expired_date,pro.sales_price, pro.alternate_price, pro.product_category_id,pc.name AS product_category_name,un.name AS unit_name,pro.branch_id").
		Joins("LEFT JOIN product_categories pc ON pc.id = pro.product_category_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("pro.id = ?", id).
		Scan(&allProduct).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Data tidak ditemukan", err)
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "Data ditemukan", allProduct)
}

func (h *ProductHandler) GetAllProduct(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	var allProduct []models.ProductDetail
	query := configs.DB.Table("products pro").
		Select("pro.id,pro.sku,pro.name, pro.alias, pro.description, pro.ingredient, pro.dosage, pro.side_affection, pro.unit_id, un.name AS unit_name,pro.stock,pro.purchase_price,pro.sales_price,pro.alternate_price,pro.expired_date, pro.product_category_id, pc.name AS product_category_name").
		Joins("LEFT JOIN product_categories pc ON pc.id = pro.product_category_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("pro.branch_id = ?", branchID)
	_, search, total, page, totalPages, limit, err := helpers.Paginate(c, query, &allProduct, []string{"pro.name ILIKE ?", "pro.alias ILIKE ?", "pro.description ILIKE ?", "pro.ingredient ILIKE ?", "pro.dosage ILIKE ?", "pro.side_affection ILIKE ?"})
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get AllProduct failed", err.Error())
	}
	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Products retrieved successfully", search, int(total), page, int(totalPages), int(limit), allProduct)
}

func (h *ProductHandler) CmbProdSale(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	search := strings.TrimSpace(c.Query("search"))
	var cmbProducts []models.ProdSaleCombo
	query := configs.DB.Table("products").
		Select("products.id as product_id, products.name as product_name, sales_price AS price, products.stock, products.unit_id, units.name AS unit_name").
		Joins("LEFT JOIN units ON units.id = products.unit_id").
		Where("products.branch_id = ?", branchID)
	search = strings.ToLower(search)
	query = query.Where("products.name ILIKE ? OR products.description ILIKE ? OR products.id ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	query = query.Order("products.name ASC")
	if err := query.Scan(&cmbProducts).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get Combo Products failed", err)
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "Combo Products retrieved successfully", cmbProducts)
}

func (h *ProductHandler) CmbProdPurchase(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	search := strings.TrimSpace(c.Query("search"))
	var cmbProducts []models.ProdPurchaseCombo
	query := configs.DB.Table("products").
		Select("products.id as product_id, products.name as product_name, purchase_price AS price, products.unit_id, units.name AS unit_name").
		Joins("LEFT JOIN units ON units.id = products.unit_id").
		Where("products.branch_id = ?", branchID)
	search = strings.ToLower(search)
	query = query.Where("products.name ILIKE ? OR products.description ILIKE ? OR products.id ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	query = query.Order("products.name ASC")
	if err := query.Scan(&cmbProducts).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get Combo Purchase Products failed", err)
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "Combo Purchase Products retrieved successfully", cmbProducts)
}
