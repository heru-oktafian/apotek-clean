package handlers

import (
	math "math"
	strconv "strconv"
	strings "strings"
	time "time"

	configs "apotek-clean/configs"
	helpers "apotek-clean/helpers"
	models "apotek-clean/models"
	services "apotek-clean/services"
	fiber "github.com/gofiber/fiber/v2"
)

type DefectaHandler struct{}

func NewDefectaHandler() *DefectaHandler {
	return &DefectaHandler{}
}

func (h *DefectaHandler) CreateDefecta(c *fiber.Ctx) error {
	nowWIB := time.Now().In(configs.Location)
	branchID, _ := services.GetBranchID(c)
	generatedID := helpers.GenerateID("DFT")
	var input models.DefectaInput
	if err := c.BodyParser(&input); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid Input", nil)
	}
	parsedDate, err := services.ParseDefectaDate(input.DefectaDate, nowWIB)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD", nil)
	}
	defecta := models.Defectas{ID: generatedID, DefectaDate: parsedDate, TotalEstimate: 0, DefectaStatus: models.Active, BranchID: branchID, CreatedAt: nowWIB, UpdatedAt: nowWIB}
	if err := configs.DB.Create(&defecta).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to create defecta", nil)
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "Defecta created successfully", defecta)
}

func (h *DefectaHandler) UpdateDefecta(c *fiber.Ctx) error {
	nowWIB := time.Now().In(configs.Location)
	id := c.Params("id")
	var input models.DefectaInput
	var defecta models.Defectas
	if err := c.BodyParser(&input); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid input", err)
	}
	if err := configs.DB.First(&defecta, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Defecta not found", nil)
	}
	if input.DefectaDate != "" {
		parsedDate, err := services.ParseDefectaDate(input.DefectaDate, nowWIB)
		if err != nil {
			return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD", nil)
		}
		defecta.DefectaDate = parsedDate
	}
	if input.DefectaStatus != "" {
		defecta.DefectaStatus = input.DefectaStatus
	}
	defecta.DefectaStatus = input.DefectaStatus
	defecta.UpdatedAt = nowWIB
	if err := configs.DB.Save(&defecta).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to update defecta", nil)
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "Defecta updated successfully", defecta)
}

func (h *DefectaHandler) DeleteDefecta(c *fiber.Ctx) error {
	id := c.Params("id")
	var defecta models.Defectas
	if err := configs.DB.First(&defecta, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Defecta not found", nil)
	}
	if err := configs.DB.Where("defecta_id = ?", defecta.ID).Delete(&models.DefectaItems{}).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to delete defecta items", nil)
	}
	if err := configs.DB.Delete(&defecta).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to delete defecta", nil)
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "Defecta deleted successfully", nil)
}

func (h *DefectaHandler) CreateDefectaItem(c *fiber.Ctx) error {
	var input models.DefectaInputItem
	if err := c.BodyParser(&input); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid Input", nil)
	}
	generatedID := helpers.GenerateID("DFI")
	var existingItem models.DefectaItems
	result := configs.DB.Where("defecta_id = ? AND product_id = ?", input.DefectaId, input.ProductId).First(&existingItem)
	if result.Error == nil {
		existingItem.Qty += input.Qty
		existingItem.SubTotal = services.SumDefectaSubTotal(existingItem.Price, existingItem.Qty)
		if err := configs.DB.Save(&existingItem).Error; err != nil {
			return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to update defecta item", nil)
		}
		var totalEstimate int
		configs.DB.Table("defecta_items").Where("defecta_id = ?", existingItem.DefectaId).Select("COALESCE(SUM(sub_total), 0)").Row().Scan(&totalEstimate)
		configs.DB.Table("defectas").Where("id = ?", existingItem.DefectaId).Update("total_estimate", totalEstimate)
		return helpers.JSONResponse(c, fiber.StatusOK, "Defecta item updated successfully", existingItem)
	}
	defectaItem := models.DefectaItems{ID: generatedID, DefectaId: input.DefectaId, ProductId: input.ProductId, UnitId: input.UnitId, Price: input.Price, Qty: input.Qty, SubTotal: services.SumDefectaSubTotal(input.Price, input.Qty)}
	if err := configs.DB.Create(&defectaItem).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to create defecta item", nil)
	}
	var totalEstimate int
	configs.DB.Table("defecta_items").Where("defecta_id = ?", defectaItem.DefectaId).Select("COALESCE(SUM(sub_total), 0)").Row().Scan(&totalEstimate)
	configs.DB.Table("defectas").Where("id = ?", defectaItem.DefectaId).Update("total_estimate", totalEstimate)
	return helpers.JSONResponse(c, fiber.StatusOK, "Defecta item created successfully", defectaItem)
}

func (h *DefectaHandler) UpdateDefectaItem(c *fiber.Ctx) error {
	id := c.Params("id")
	var input models.DefectaInputItem
	if err := c.BodyParser(&input); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid Input", nil)
	}
	var defectaItem models.DefectaItems
	if err := configs.DB.First(&defectaItem, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Defecta item not found", nil)
	}
	defectaItem.ProductId = input.ProductId
	if input.UnitId != "" {
		defectaItem.UnitId = input.UnitId
	}
	if input.Price != 0 {
		defectaItem.Price = input.Price
	}
	defectaItem.Qty = input.Qty
	defectaItem.SubTotal = services.SumDefectaSubTotal(input.Price, input.Qty)
	if err := configs.DB.Save(&defectaItem).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to update defecta item", nil)
	}
	var totalEstimate int
	configs.DB.Table("defecta_items").Where("defecta_id = ?", defectaItem.DefectaId).Select("COALESCE(SUM(sub_total), 0)").Row().Scan(&totalEstimate)
	configs.DB.Table("defectas").Where("id = ?", defectaItem.DefectaId).Update("total_estimate", totalEstimate)
	return helpers.JSONResponse(c, fiber.StatusOK, "Defecta item updated successfully", defectaItem)
}

func (h *DefectaHandler) DeleteDefectaItem(c *fiber.Ctx) error {
	id := c.Params("id")
	var defectaItem models.DefectaItems
	if err := configs.DB.First(&defectaItem, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Defecta item not found", nil)
	}
	if err := configs.DB.Delete(&defectaItem).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to delete defecta item", nil)
	}
	var totalEstimate int
	configs.DB.Table("defecta_items").Where("defecta_id = ?", defectaItem.DefectaId).Select("COALESCE(SUM(sub_total), 0)").Row().Scan(&totalEstimate)
	configs.DB.Table("defectas").Where("id = ?", defectaItem.DefectaId).Update("total_estimate", totalEstimate)
	return helpers.JSONResponse(c, fiber.StatusOK, "Defecta item deleted successfully", nil)
}

func (h *DefectaHandler) GetAllDefectas(c *fiber.Ctx) error {
	nowWIB := time.Now().In(configs.Location)
	branchID, _ := services.GetBranchID(c)
	pageParam := c.Query("page")
	search := strings.TrimSpace(c.Query("search"))
	page := 1
	if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
		page = p
	}
	limit := 10
	offset := (page - 1) * limit
	month := strings.TrimSpace(c.Query("month"))
	if month == "" {
		month = nowWIB.Format("2006-01")
	}
	var defectas []models.Defectas
	var total int64
	query := configs.DB.Table("defectas df").Select("df.id, df.defecta_date, df.total_estimate, df.defecta_status").Where("df.branch_id = ?", branchID)
	startDate, err := time.Parse("2006-01", month)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid month format. Use YYYY-MM", nil)
	}
	endDate := startDate.AddDate(0, 1, 0)
	query = query.Where("df.defecta_date >= ? AND df.defecta_date < ?", startDate, endDate)
	if search != "" {
		likeSearch := "%" + search + "%"
		query = query.Where("df.id LIKE ? OR df.defecta_status LIKE ?", likeSearch, likeSearch)
	}
	if err := query.Count(&total).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to count defectas", nil)
	}
	if err := query.Order("df.created_at DESC").Limit(limit).Offset(offset).Find(&defectas).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to fetch defectas", nil)
	}
	totalPages, formattedDefectas := services.BuildDefectaListResponse(defectas, limit, total, page)
	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Defectas retrieved successfully", search, int(total), page, totalPages, limit, formattedDefectas)
}

func (h *DefectaHandler) GetAllDefectaItems(c *fiber.Ctx) error {
	defectaID := c.Params("id")
	var defectaItems []models.AllDefectaItems
	query := configs.DB.Table("defecta_items di").
		Select("di.id, di.defecta_id, pro.name as product_name, un.name as unit_name, di.price, di.qty, di.sub_total").
		Joins("LEFT JOIN products pro ON pro.id = di.product_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("di.defecta_id = ?", defectaID)
	if err := query.Find(&defectaItems).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to fetch defecta items", nil)
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "Defecta items retrieved successfully", defectaItems)
}

func (h *DefectaHandler) GetDefetaWithItems(c *fiber.Ctx) error {
	defectaID := c.Params("id")
	var defecta models.Defectas
	if err := configs.DB.First(&defecta, "id = ?", defectaID).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Defecta not found", nil)
	}
	var defectaItems []models.AllDefectaItems
	if err := configs.DB.Table("defecta_items di").
		Select("di.id, di.product_id, pro.name as product_name, di.unit_id, un.name as unit_name, di.price, di.qty, di.sub_total").
		Joins("LEFT JOIN products pro ON pro.id = di.product_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("di.defecta_id = ?", defecta.ID).
		Find(&defectaItems).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to fetch defecta items", nil)
	}
	formattedDefectaItems := services.BuildDefectaItemsResponse(defectaItems)
	response := models.DefectaDetailWithItemsResponse{ID: defecta.ID, DefectaDate: helpers.FormatIndonesianDate(defecta.DefectaDate), TotalEstimate: defecta.TotalEstimate, DefectaStatus: string(defecta.DefectaStatus), Items: formattedDefectaItems}
	return helpers.JSONResponse(c, fiber.StatusOK, "Defecta details retrieved successfully", response)
}
