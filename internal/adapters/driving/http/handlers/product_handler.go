package handlers

import (
	"strings"

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
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal membuat produk", err.Error())
	}
	return helpers.JSONResponse(c, fiber.StatusCreated, "Produk berhasil dibuat", product)
}

func (h *ProductHandler) GetProduct(c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "ID tidak boleh kosong", nil)
	}
	var allProduct []models.ProductDetail
	if err := configs.DB.
		Table("products pro").
		Select("pro.id,pro.sku,pro.name,pro.description, pro.ingredient, pro.dosage, pro.side_affection, pro.unit_id AS unit_id,pro.stock,pro.showcase_stock,pro.warehouse_stock,pro.purchase_price,pro.expired_date,pro.sales_price, pro.alternate_price, pro.product_category_id,pc.name AS product_category_name,un.name AS unit_name,pro.branch_id").
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

	// Count query — Paginate mutates its query, so keep it separate from data query
	countQuery := configs.DB.Table("products pro").
		Select("pro.id").
		Joins("LEFT JOIN product_categories pc ON pc.id = pro.product_category_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("pro.branch_id = ?", branchID)

	// Paginate only for pagination metadata; uses dummy model to avoid premature data scan
	_, search, total, page, totalPages, limit, err := helpers.Paginate(c, countQuery, new([]models.ProductDetail), []string{"pro.name ILIKE ?", "pro.alias ILIKE ?", "pro.description ILIKE ?", "pro.ingredient ILIKE ?", "pro.dosage ILIKE ?", "pro.side_affection ILIKE ?"})
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get AllProduct failed", err.Error())
	}

	// Data query — fresh instance with full SELECT + ORDER BY (separate from countQuery so ORDER survives)
	dataQuery := configs.DB.Table("products pro").
		Select("pro.id,pro.sku,pro.name, pro.alias, pro.description, pro.ingredient, pro.dosage, pro.side_affection, pro.unit_id, un.name AS unit_name,pro.stock,pro.showcase_stock,pro.warehouse_stock,pro.purchase_price,pro.sales_price,pro.alternate_price,pro.expired_date, pro.product_category_id, pc.name AS product_category_name").
		Joins("LEFT JOIN product_categories pc ON pc.id = pro.product_category_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("pro.branch_id = ?", branchID)

	// Apply same search filters that Paginate uses
	if search != "" {
		searchLower := strings.ToLower(search)
		dataQuery = dataQuery.Where(
			"pro.name ILIKE ? OR pro.alias ILIKE ? OR pro.description ILIKE ? OR pro.ingredient ILIKE ? OR pro.dosage ILIKE ? OR pro.side_affection ILIKE ?",
			"%"+searchLower+"%", "%"+searchLower+"%", "%"+searchLower+"%", "%"+searchLower+"%", "%"+searchLower+"%", "%"+searchLower+"%",
		)
	}

	offset := (page - 1) * limit
	err = dataQuery.Order("pro.name ASC").Offset(offset).Limit(limit).Find(&allProduct).Error
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get AllProduct failed", err.Error())
	}

	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Products retrieved successfully", search, int(total), page, totalPages, limit, allProduct)
}

func (h *ProductHandler) CmbProdSale(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	search := strings.TrimSpace(c.Query("search"))
	var cmbProducts []models.ProdSaleCombo
	query := configs.DB.Table("products").
		Select("products.id as product_id, products.name as product_name, sales_price AS price, products.stock, products.showcase_stock, products.warehouse_stock, products.unit_id, units.name AS unit_name").
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

func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	var product models.Product
	if err := configs.DB.Where("id = ?", id).First(&product).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Produk tidak ditemukan", err)
	}
	var newProduct models.Product
	if err := c.BodyParser(&newProduct); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid input", err)
	}
	product.Name = newProduct.Name
	product.SKU = newProduct.SKU
	product.Alias = newProduct.Alias
	product.Description = newProduct.Description
	product.Ingredient = newProduct.Ingredient
	product.Dosage = newProduct.Dosage
	product.SideAffection = newProduct.SideAffection
	product.UnitId = newProduct.UnitId
	product.ProductCategoryId = newProduct.ProductCategoryId
	product.PurchasePrice = newProduct.PurchasePrice
	product.SalesPrice = newProduct.SalesPrice
	product.AlternatePrice = newProduct.AlternatePrice
	product.ExpiredDate = newProduct.ExpiredDate
	if err := configs.DB.Save(&product).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal update produk", err.Error())
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "Produk berhasil diupdate", product)
}

func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := configs.DB.Delete(&models.Product{}, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal hapus produk", err.Error())
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "Produk berhasil dihapus", nil)
}
