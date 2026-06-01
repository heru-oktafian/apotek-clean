package controllers

import (
	strings "strings"

	fiber "github.com/gofiber/fiber/v2"
	config "github.com/heru-oktafian/fiber-apotek/configs"
	helpers "github.com/heru-oktafian/fiber-apotek/helpers"
	models "github.com/heru-oktafian/fiber-apotek/models"
	services "github.com/heru-oktafian/fiber-apotek/services"
)

// CreateProductCategory buat product category
func CreateProductCategory(c *fiber.Ctx) error {
	// Membuat ProductCategory baru menggunakan helpers
	return helpers.CreateResourceInc(c, config.DB, &models.ProductCategory{})
}

// UpdateProductCategory update ProductCategory
func UpdateProductCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	// Memperbarui ProductCategory menggunakan helpers
	return helpers.UpdateResource(c, config.DB, &models.ProductCategory{}, id)
}

// DeleteProductCategory hapus ProductCategory
func DeleteProductCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	// Menghapus ProductCategory menggunakan helpers
	return helpers.DeleteResource(c, config.DB, &models.ProductCategory{}, id)
}

// GetProductCategory tampilkan ProductCategory berdasarkan id
func GetProductCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	// Mengambil ProductCategory menggunakan helpers
	return helpers.GetResource(c, config.DB, &models.ProductCategory{}, id)
}

// CmbProductCategory mendapatkan semua kategori produk
func CmbProductCategory(c *fiber.Ctx) error {
	// Get branch id
	branch_id, _ := services.GetBranchID(c)

	// Parsing query parameter "search"
	search := strings.TrimSpace(c.Query("search"))

	var categories []models.ComboProductCategory

	// Query untuk mendapatkan semua kategori produk
	query := config.DB.Table("product_categories").
		Select("product_categories.id as product_category_id, product_categories.name as product_category_name").
		Where("branch_id = ?", branch_id)

	// Jika ada search key, tambahkan filter WHERE
	if search != "" {
		search = strings.ToLower(search) // Konversi search ke lowercase
		query = query.Where("LOWER(product_categories.name) LIKE ?", "%"+search+"%")
	}

	// Tambahkan urutan ascending berdasarkan nama
	query = query.Order("product_categories.name ASC")

	// Eksekusi query
	if err := query.Find(&categories).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to get data", "Failed to get data")
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Data berhasil ditemukan", categories)
}

// GetAllProductCategory tampilkan semua ProductCategory
func GetAllProductCategory(c *fiber.Ctx) error {
	// Get branch id
	branch_id, _ := services.GetBranchID(c)

	// Inisialisasi slice untuk menampung kategori produk
	var ProductCategory []models.ComboProductCategory

	// Query dasar
	query := config.DB.Table("product_categories pc").Select("pc.id AS product_category_id, pc.name AS product_category_name").Where("pc.branch_id = ?", branch_id)

	// Hitung total data untuk pagination
	_, search, total, page, totalPages, limit, err := helpers.Paginate(c, query, &ProductCategory, []string{"pc.name ILIKE ?"})
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get Product Categories failed", err.Error())
	}

	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Product Categories retrieved successfully", search, total, page, totalPages, limit, ProductCategory)
}
