package handlers

import (
	fmt "fmt"
	strings "strings"

	configs "apotek-clean/configs"
	helpers "apotek-clean/helpers"
	models "apotek-clean/internal/core/entities"
	services "apotek-clean/services"
	fiber "github.com/gofiber/fiber/v2"
	gorm "gorm.io/gorm"
)

type UnitConversionHandler struct{}

func NewUnitConversionHandler() *UnitConversionHandler {
	return &UnitConversionHandler{}
}

func (h *UnitConversionHandler) CreateUnitConversion(c *fiber.Ctx) error {
	db := configs.DB
	var req models.UnitConversionRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Format data yang dikirim tidak valid", err)
	}
	branchID, _ := services.GetBranchID(c)
	if branchID == "" {
		return helpers.JSONResponse(c, fiber.StatusUnauthorized, "Branch ID tidak ditemukan di token. Unauthorized.", nil)
	}
	tx := db.Begin()
	if tx.Error != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal memulai transaksi database", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	var existingConversion models.UnitConversion
	checkErr := tx.Where("product_id = ? AND init_id = ? AND final_id = ? AND branch_id = ?", req.ProductId, req.InitId, req.FinalId, branchID).First(&existingConversion).Error
	if checkErr == nil {
		tx.Rollback()
		return helpers.JSONResponse(c, fiber.StatusConflict, fmt.Sprintf("Konversi satuan dari %s ke %s untuk produk %s sudah ada di cabang ini", req.InitId, req.FinalId, req.ProductId), nil)
	} else if checkErr != gorm.ErrRecordNotFound {
		tx.Rollback()
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal memeriksa konversi satuan yang sudah ada", checkErr)
	}
	var product models.Product
	if err := tx.Where("id = ? AND branch_id = ?", req.ProductId, branchID).First(&product).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return helpers.JSONResponse(c, fiber.StatusNotFound, fmt.Sprintf("Produk dengan ID %s tidak ditemukan di cabang %s", req.ProductId, branchID), nil)
		}
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil produk untuk validasi", err)
	}
	var initUnit models.Unit
	if err := tx.Where("id = ? AND branch_id = ?", req.InitId, branchID).First(&initUnit).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return helpers.JSONResponse(c, fiber.StatusNotFound, fmt.Sprintf("Satuan awal (InitId) dengan ID %s tidak ditemukan di cabang %s", req.InitId, branchID), nil)
		}
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil satuan awal untuk validasi", err)
	}
	var finalUnit models.Unit
	if err := tx.Where("id = ? AND branch_id = ?", req.FinalId, branchID).First(&finalUnit).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return helpers.JSONResponse(c, fiber.StatusNotFound, fmt.Sprintf("Satuan akhir (FinalId) dengan ID %s tidak ditemukan di cabang %s", req.FinalId, branchID), nil)
		}
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil satuan akhir untuk validasi", err)
	}
	unitConversion := models.UnitConversion{ID: helpers.GenerateID("UNC"), ProductId: req.ProductId, InitId: req.InitId, FinalId: req.FinalId, ValueConv: req.ValueConv, BranchID: branchID}
	if err := tx.Create(&unitConversion).Error; err != nil {
		tx.Rollback()
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal membuat konversi satuan", err)
	}
	if err := tx.Commit().Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal melakukan commit transaksi database", err)
	}
	return helpers.JSONResponse(c, fiber.StatusCreated, "Konversi satuan berhasil dibuat", fiber.Map{"id": unitConversion.ID, "product_id": unitConversion.ProductId, "product_name": product.Name, "init_id": unitConversion.InitId, "init_unit_name": initUnit.Name, "final_id": unitConversion.FinalId, "final_unit_name": finalUnit.Name, "value_conv": unitConversion.ValueConv, "branch_id": unitConversion.BranchID})
}

func (h *UnitConversionHandler) UpdateUnitConversion(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.UpdateResource(c, configs.DB, &models.UnitConversion{}, id)
}

func (h *UnitConversionHandler) DeleteUnitConversion(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.DeleteResource(c, configs.DB, &models.UnitConversion{}, id)
}

func (h *UnitConversionHandler) GetUnitConversionByID(c *fiber.Ctx) error {
	id := c.Params("id")
	return helpers.GetResource(c, configs.DB, &models.UnitConversion{}, id)
}

func (h *UnitConversionHandler) GetAllUnitConversion(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	var unitConversions []models.UnitConversionDetail
	query := configs.DB.Table("unit_conversions unc").
		Select("unc.id, pro.name AS product_name, uin.name AS init_name, ufi.name AS final_name, unc.value_conv, unc.product_id, unc.init_id, unc.final_id, unc.branch_id").
		Joins("INNER JOIN products pro ON pro.id = unc.product_id AND pro.branch_id = ?", branchID).
		Joins("INNER JOIN units uin ON uin.id = unc.init_id AND uin.branch_id = ?", branchID).
		Joins("INNER JOIN units ufi ON ufi.id = unc.final_id AND ufi.branch_id = ?", branchID).
		Where("unc.branch_id = ?", branchID)
	_, search, total, page, totalPages, limit, err := helpers.Paginate(c, query, &unitConversions, []string{"pro.name ILIKE ?", "uin.name ILIKE ?", "ufi.name ILIKE ?"})
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data konversi satuan", err.Error())
	}
	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Data konversi satuan berhasil diambil", search, int(total), page, int(totalPages), int(limit), unitConversions)
}

func (h *UnitConversionHandler) CmbProdConv(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	search := strings.TrimSpace(c.Query("search"))
	var cmbProducts []models.ProdConvCombo
	query := configs.DB.Table("products").Select("id as product_id, name as product_name").Where("branch_id = ?", branchID).Order("name ASC")
	if search != "" {
		search = strings.ToLower(search)
		query = query.Where("LOWER(name) LIKE ?", "%"+search+"%")
	}
	if err := query.Find(&cmbProducts).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data", "Gagal mengambil data")
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "Data berhasil ditemukan", cmbProducts)
}
