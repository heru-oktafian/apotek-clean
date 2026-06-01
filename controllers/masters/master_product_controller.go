package controllers

import (
	strings "strings"

	fiber "github.com/gofiber/fiber/v2"
	configs "github.com/heru-oktafian/fiber-apotek/configs"
	helpers "github.com/heru-oktafian/fiber-apotek/helpers"
	models "github.com/heru-oktafian/fiber-apotek/models"
	services "github.com/heru-oktafian/fiber-apotek/services"
)

// CreateProduct buat Product
func CreateProduct(c *fiber.Ctx) error {
	var product models.Product

	if err := c.BodyParser(&product); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid input", err)
	}

	// Generate ID untuk product
	product.ID = helpers.GenerateID("PRD")

	// Set BranchID
	branchID, _ := services.GetBranchID(c)
	product.BranchID = branchID

	// Set Stock default
	product.Stock = 0

	// Jika SKU kosong atau hanya berisi spasi, samakan dengan ID produk
	if strings.TrimSpace(product.SKU) == "" {
		product.SKU = product.ID
	}

	// Simpan ke database
	if err := configs.DB.Create(&product).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to create resource", err)
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Resource created successfully", &product)
}

// UpdateProduct update Product
func UpdateProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	// Updating Product using helpers
	return helpers.UpdateResource(c, configs.DB, &models.Product{}, id)
}

// DeleteProduct hapus Product
func DeleteProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	// Deleting Product using helpers
	return helpers.DeleteResource(c, configs.DB, &models.Product{}, id)
}

// GetProduct tampilkan Product berdasarkan id
func GetProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	var AllProduct []models.ProductDetail
	if err := configs.DB.
		Table("products pro").
		Select("pro.id,pro.sku,pro.name,pro.description, pro.ingredient, pro.dosage, pro.side_affection, pro.unit_id AS unit_id,pro.stock,pro.purchase_price,pro.expired_date,pro.sales_price, pro.alternate_price, pro.product_category_id,pc.name AS product_category_name,un.name AS unit_name,pro.branch_id").
		Joins("LEFT JOIN product_categories pc ON pc.id = pro.product_category_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("pro.id = ?", id).
		Scan(&AllProduct).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Data tidak ditemukan", err)
	}

	// print(AllProduct)
	return helpers.JSONResponse(c, fiber.StatusOK, "Data ditemukan", AllProduct)
}

// GetAllProduct tampilkan semua Product
func GetAllProduct(c *fiber.Ctx) error {
	// Ambil ID cabang
	branch_id, _ := services.GetBranchID(c)

	var AllProduct []models.ProductDetail
	var total int

	// Query dasar
	query := configs.DB.Debug().Table("products pro").
		Select("pro.id,pro.sku,pro.name, pro.alias, pro.description, pro.ingredient, pro.dosage, pro.side_affection, pro.unit_id, un.name AS unit_name,pro.stock,pro.purchase_price,pro.sales_price,pro.alternate_price,pro.expired_date, pro.product_category_id, pc.name AS product_category_name").
		Joins("LEFT JOIN product_categories pc ON pc.id = pro.product_category_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("pro.branch_id = ?", branch_id)

	_, search, total, page, totalPages, limit, err := helpers.Paginate(c, query, &AllProduct, []string{"pro.name ILIKE ?", "pro.alias ILIKE ?", "pro.description ILIKE ?", "pro.ingredient ILIKE ?", "pro.dosage ILIKE ?", "pro.side_affection ILIKE ?"})
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get AllProduct failed", err.Error())
	}

	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Products retrieved successfully", search, int(total), page, int(totalPages), int(limit), AllProduct)

}

// CmbProdSale mengembalikan daftar produk untuk combo box transaksi penjualan
func CmbProdSale(c *fiber.Ctx) error {
	branch_id, _ := services.GetBranchID(c)
	search := strings.TrimSpace(c.Query("search"))

	var cmbProducts []models.ProdSaleCombo

	query := configs.DB.Table("products").
		Select("products.id as product_id, products.name as product_name, sales_price AS price, products.stock, products.unit_id, units.name AS unit_name").
		Joins("LEFT JOIN units ON units.id = products.unit_id").
		Where("products.branch_id = ?", branch_id)

	search = strings.ToLower(search)
	query = query.Where("products.name ILIKE ? OR products.description ILIKE ? OR products.id ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")

	query = query.Order("products.name ASC")

	if err := query.Scan(&cmbProducts).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get Combo Products failed", err)
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Combo Products retrieved successfully", cmbProducts)
}

// CmbProdPurchase mengembalikan daftar produk untuk combo box transaksi pembelian
func CmbProdPurchase(c *fiber.Ctx) error {
	branch_id, _ := services.GetBranchID(c)
	search := strings.TrimSpace(c.Query("search"))

	var cmbProducts []models.ProdPurchaseCombo

	query := configs.DB.Table("products").
		Select("products.id as product_id, products.name as product_name, purchase_price AS price, products.unit_id, units.name AS unit_name").
		Joins("LEFT JOIN units ON units.id = products.unit_id").
		Where("products.branch_id = ?", branch_id)

	search = strings.ToLower(search)
	query = query.Where("products.name ILIKE ? OR products.description ILIKE ? OR products.id ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")

	query = query.Order("products.name ASC")

	if err := query.Scan(&cmbProducts).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get Combo Purchase Products failed", err)
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Combo Purchase Products retrieved successfully", cmbProducts)
}
