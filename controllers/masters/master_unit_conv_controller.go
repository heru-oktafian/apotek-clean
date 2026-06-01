package controllers

import (
	fmt "fmt"
	strings "strings"

	fiber "github.com/gofiber/fiber/v2"
	configs "github.com/heru-oktafian/fiber-apotek/configs"
	helpers "github.com/heru-oktafian/fiber-apotek/helpers"
	models "github.com/heru-oktafian/fiber-apotek/models"
	services "github.com/heru-oktafian/fiber-apotek/services"
	gorm "gorm.io/gorm"
)

// CreateUnitConversion controller
// Endpoint: POST /api/unit-conversions
func CreateUnitConversion(c *fiber.Ctx) error {
	db := configs.DB
	var req models.UnitConversionRequest

	// Parsing request body
	if err := c.BodyParser(&req); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Format data yang dikirim tidak valid", err)
	}

	// Mendapatkan BranchID dari token (asumsi UnitConversion spesifik per cabang)
	branchID, _ := services.GetBranchID(c)
	if branchID == "" {
		return helpers.JSONResponse(c, fiber.StatusUnauthorized, "Branch ID not found in token. Unauthorized", nil)
	}

	// --- VALIDASI INPUT ---
	// if err := helpers.ValidateStruct(req); err != nil {
	// 	return responses.BadRequest(c, "Validation failed", err)
	// }
	// --- AKHIR VALIDASI INPUT ---

	// Mulai transaksi database
	tx := db.Begin()
	if tx.Error != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to begin database transaction", tx.Error)
	}
	// Pastikan transaksi di-rollback jika terjadi kesalahan atau panic
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// --- Pengecekan Duplikasi ---
	var existingConversion models.UnitConversion
	checkErr := tx.Where("product_id = ? AND init_id = ? AND final_id = ? AND branch_id = ?",
		req.ProductId,
		req.InitId,
		req.FinalId,
		branchID,
	).First(&existingConversion).Error

	if checkErr == nil {
		// Jika record sudah ditemukan (checkErr == nil), berarti duplikat
		tx.Rollback()
		return helpers.JSONResponse(c, fiber.StatusConflict, fmt.Sprintf("unit conversion from '%s' to '%s' for product '%s' already exists in this branch: duplicate entry",
			req.InitId, req.FinalId, req.ProductId), nil)
	} else if checkErr != gorm.ErrRecordNotFound {
		// Jika ada error lain selain record not found
		tx.Rollback()
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to check for existing unit conversion", checkErr)
	}
	// Jika checkErr == gorm.ErrRecordNotFound, berarti tidak ada duplikat, lanjutkan proses

	// --- 1. Dapatkan Product untuk memastikan ProductId valid ---
	var product models.Product
	if err := tx.Where("id = ? AND branch_id = ?", req.ProductId, branchID).First(&product).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return helpers.JSONResponse(c, fiber.StatusNotFound, fmt.Sprintf("Product with ID %s not found in branch %s.", req.ProductId, branchID), nil)
		}
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to retrieve product for validation", err)
	}

	// --- 2. Dapatkan Init Unit untuk memastikan InitId valid ---
	var initUnit models.Unit
	if err := tx.Where("id = ? AND branch_id = ?", req.InitId, branchID).First(&initUnit).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return helpers.JSONResponse(c, fiber.StatusNotFound, fmt.Sprintf("Initial unit (InitId) with ID %s not found in branch %s.", req.InitId, branchID), nil)
		}
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to retrieve initial unit for validation", err)
	}

	// --- 3. Dapatkan Final Unit untuk memastikan FinalId valid ---
	var finalUnit models.Unit
	if err := tx.Where("id = ? AND branch_id = ?", req.FinalId, branchID).First(&finalUnit).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return helpers.JSONResponse(c, fiber.StatusNotFound, fmt.Sprintf("Final unit (FinalId) with ID %s not found in branch %s.", req.FinalId, branchID), nil)
		}
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to retrieve final unit for validation", err)
	}

	// --- Buat objek UnitConversion baru ---
	unitConversion := models.UnitConversion{
		ID:        helpers.GenerateID("UNC"), // Generate ID untuk Unit Conversion
		ProductId: req.ProductId,
		InitId:    req.InitId,
		FinalId:   req.FinalId,
		ValueConv: req.ValueConv,
		BranchID:  branchID, // Set BranchID dari token
	}

	// Simpan UnitConversion ke database
	if err := tx.Create(&unitConversion).Error; err != nil {
		tx.Rollback()
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to create unit conversion", err)
	}

	// Commit transaksi jika semua berhasil
	if err := tx.Commit().Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to commit database transaction", err)
	}

	// Berhasil
	return helpers.JSONResponse(c, fiber.StatusCreated, "Unit conversion created successfully", fiber.Map{
		"id":              unitConversion.ID,
		"product_id":      unitConversion.ProductId,
		"product_name":    product.Name, // Menambahkan nama produk
		"init_id":         unitConversion.InitId,
		"init_unit_name":  initUnit.Name, // Menambahkan nama unit
		"final_id":        unitConversion.FinalId,
		"final_unit_name": finalUnit.Name, // Menambahkan nama unit
		"value_conv":      unitConversion.ValueConv,
		"branch_id":       unitConversion.BranchID,
	})
}

// UpdateUnit update unit
func UpdateUnitConversion(c *fiber.Ctx) error {
	id := c.Params("id")
	// Updating unit using helpers
	return helpers.UpdateResource(c, configs.DB, &models.UnitConversion{}, id)
}

// DeleteUnit hapus unit
func DeleteUnitConversion(c *fiber.Ctx) error {
	id := c.Params("id")
	// Deleting unit using helpers
	return helpers.DeleteResource(c, configs.DB, &models.UnitConversion{}, id)
}

// GetUnitConversionByID tampilkan unit berdasarkan id
func GetUnitConversionByID(c *fiber.Ctx) error {
	id := c.Params("id")
	// Getting unit using helpers
	return helpers.GetResource(c, configs.DB, &models.UnitConversion{}, id)
}

// GetAllUnit tampilkan semua unit
func GetAllUnitConversion(c *fiber.Ctx) error {
	// Ambil ID cabang
	branch_id, _ := services.GetBranchID(c)

	var unit_conversions []models.UnitConversionDetail

	// Query dasar
	query := configs.DB.Table("unit_conversions unc").
		Select("unc.id, pro.name AS product_name, uin.name AS init_name, ufi.name AS final_name, unc.value_conv, unc.product_id, unc.init_id, unc.final_id, unc.branch_id").
		Joins("INNER JOIN products pro ON pro.id = unc.product_id AND pro.branch_id = ?", branch_id).
		Joins("INNER JOIN units uin ON uin.id = unc.init_id AND uin.branch_id = ?", branch_id).
		Joins("INNER JOIN units ufi ON ufi.id = unc.final_id AND ufi.branch_id = ?", branch_id).
		Where("unc.branch_id = ?", branch_id)

	_, search, total, page, totalPages, limit, err := helpers.Paginate(c, query, &unit_conversions, []string{"pro.name ILIKE ?", "uin.name ILIKE ?", "ufi.name ILIKE ?"})
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get Unit Conversions failed", err.Error())
	}

	// Return response
	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Unit conversions retrieved successfully", search, int(total), page, int(totalPages), int(limit), unit_conversions)

}

// CmbProdConv mendapatkan semua produk
func CmbProdConv(c *fiber.Ctx) error {
	// Get branch id
	branch_id, _ := services.GetBranchID(c)

	// Parsing query parameter "search"
	search := strings.TrimSpace(c.Query("search"))

	var cmbProducts []models.ProdConvCombo

	// Query untuk mendapatkan semua produk
	query := configs.DB.Table("products").
		Select("id as product_id, name as product_name").
		Where("branch_id = ?", branch_id)

	// Jika ada parameter search, tambahkan filter WHERE
	if search != "" {
		search = strings.ToLower(search) // Konversi search ke lowercase
		query = query.Where("LOWER(name) LIKE ?", "%"+search+"%")
	}

	if err := query.Find(&cmbProducts).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to get data", "Failed to get data")
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Data berhasil ditemukan", cmbProducts)
}
