package controllers

import (
	math "math"
	strconv "strconv"
	strings "strings"
	time "time"

	fiber "github.com/gofiber/fiber/v2"
	configs "github.com/heru-oktafian/fiber-apotek/configs"
	helpers "github.com/heru-oktafian/fiber-apotek/helpers"
	models "github.com/heru-oktafian/fiber-apotek/models"
	services "github.com/heru-oktafian/fiber-apotek/services"
)

// CreateDefecta handles the creation of a new defecta.
func CreateDefecta(c *fiber.Ctx) error {

	// Get the current time in WIB (Western Indonesia Time)
	nowWIB := time.Now().In(configs.Location)

	// Ambil informasi dari token melalui middleware
	branchID, _ := services.GetBranchID(c)
	// userID, _ := middlewares.GetUserID(c.Request)
	generatedID := helpers.GenerateID("DFT")

	var input models.DefectaInput
	if err := c.BodyParser(&input); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid Input", nil)
	}

	layout := "2006-01-02"
	parsedDate, err := time.Parse(layout, input.DefectaDate)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD", nil)
	}

	// Initialize database connection
	db := configs.DB

	// Create new defecta record
	defecta := models.Defectas{
		ID:            generatedID,
		DefectaDate:   parsedDate,
		TotalEstimate: 0,             // Akan dikalkulasi nanti
		DefectaStatus: models.Active, // Status awal adalah Active
		BranchID:      branchID,
		CreatedAt:     nowWIB,
		UpdatedAt:     nowWIB,
	}

	// Simpan defekta ke database
	if err := db.Create(&defecta).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to create defecta", nil)
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Defecta created successfully", defecta)
}

func UpdateDefecta(c *fiber.Ctx) error {

	nowWIB := time.Now().In(configs.Location)

	db := configs.DB
	id := c.Params("id")

	// Inisialisasi input dan defecta
	var input models.DefectaInput
	var defecta models.Defectas

	// Cek validasi input
	if err := c.BodyParser(&input); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid input", err)
	}

	// Cek apakah defecta dengan ID tersebut ada
	if err := db.First(&defecta, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Defecta not found", nil)
	}

	// Perbarui field defecta_date jika ada di input
	if input.DefectaDate != "" {
		layout := "2006-01-02"
		parsedDate, err := time.Parse(layout, input.DefectaDate)
		if err != nil {
			return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD", nil)
		}
		defecta.DefectaDate = parsedDate
	}

	// Perbarui field defecta_status jika ada di input
	if input.DefectaStatus != "" {
		defecta.DefectaStatus = input.DefectaStatus
	}

	// Perbarui field-field defecta
	defecta.DefectaStatus = input.DefectaStatus
	defecta.UpdatedAt = nowWIB

	// Simpan perubahan ke database
	if err := db.Save(&defecta).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to update defecta", nil)
	}

	// Kembalikan respons sukses
	return helpers.JSONResponse(c, fiber.StatusOK, "Defecta updated successfully", defecta)
}

func DeleteDefecta(c *fiber.Ctx) error {
	db := configs.DB
	id := c.Params("id")

	// Cek apakah defecta dengan ID tersebut ada
	var defecta models.Defectas
	if err := db.First(&defecta, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Defecta not found", nil)
	}

	// Hapus detail items yang terkait dengan defecta
	if err := db.Where("defecta_id = ?", defecta.ID).Delete(&models.DefectaItems{}).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to delete defecta items", nil)
	}

	// Hapus defecta dari database
	if err := db.Delete(&defecta).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to delete defecta", nil)
	}

	// Kembalikan respons sukses
	return helpers.JSONResponse(c, fiber.StatusOK, "Defecta deleted successfully", nil)
}

func CreateDefectaItem(c *fiber.Ctx) error {

	db := configs.DB

	var input models.DefectaInputItem
	if err := c.BodyParser(&input); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid Input", nil)
	}

	generatedID := helpers.GenerateID("DFI")

	// Cek apakah product_id sudah ada dalam defecta_items dengan defecta_id yang sama
	var existingItem models.DefectaItems
	result := db.Where("defecta_id = ? AND product_id = ?", input.DefectaId, input.ProductId).First(&existingItem)

	if result.Error == nil {
		// Item sudah ada, update qty
		existingItem.Qty += input.Qty
		existingItem.SubTotal = existingItem.Price * existingItem.Qty

		if err := db.Save(&existingItem).Error; err != nil {
			return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to update defecta item", nil)
		}

		// Update total_estimate
		var totalEstimate int
		db.Table("defecta_items").Where("defecta_id = ?", existingItem.DefectaId).Select("COALESCE(SUM(sub_total), 0)").Row().Scan(&totalEstimate)
		db.Table("defectas").Where("id = ?", existingItem.DefectaId).Update("total_estimate", totalEstimate)

		defectaItem := existingItem
		return helpers.JSONResponse(c, fiber.StatusOK, "Defecta item updated successfully", defectaItem)
	}

	// Item belum ada, buat item baru
	defectaItem := models.DefectaItems{
		ID:        generatedID,
		DefectaId: input.DefectaId,
		ProductId: input.ProductId,
		UnitId:    input.UnitId,
		Price:     input.Price,
		Qty:       input.Qty,
		SubTotal:  input.Price * input.Qty,
	}

	// Simpan defecta item ke database
	if err := db.Create(&defectaItem).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to create defecta item", nil)
	}

	// Update total_estimate
	var totalEstimate int
	db.Table("defecta_items").Where("defecta_id = ?", defectaItem.DefectaId).Select("COALESCE(SUM(sub_total), 0)").Row().Scan(&totalEstimate)
	db.Table("defectas").Where("id = ?", defectaItem.DefectaId).Update("total_estimate", totalEstimate)

	return helpers.JSONResponse(c, fiber.StatusOK, "Defecta item created successfully", defectaItem)
}

// UpdateDefectaItem menangani pembaruan item defecta yang sudah ada.
func UpdateDefectaItem(c *fiber.Ctx) error {
	db := configs.DB
	id := c.Params("id")

	var input models.DefectaInputItem
	if err := c.BodyParser(&input); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid Input", nil)
	}

	// Cek apakah defecta item dengan ID tersebut ada
	var defectaItem models.DefectaItems
	if err := db.First(&defectaItem, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Defecta item not found", nil)
	}

	// Perbarui field-field defecta item
	defectaItem.ProductId = input.ProductId
	if input.UnitId != "" {
		defectaItem.UnitId = input.UnitId
	}
	if input.Price != 0 {
		defectaItem.Price = input.Price
	}
	defectaItem.Qty = input.Qty
	defectaItem.SubTotal = input.Price * input.Qty

	// Simpan perubahan ke database
	if err := db.Save(&defectaItem).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to update defecta item", nil)
	}

	// Update total_estimate
	var totalEstimate int
	db.Table("defecta_items").Where("defecta_id = ?", defectaItem.DefectaId).Select("COALESCE(SUM(sub_total), 0)").Row().Scan(&totalEstimate)
	db.Table("defectas").Where("id = ?", defectaItem.DefectaId).Update("total_estimate", totalEstimate)

	// Kembalikan respons sukses
	return helpers.JSONResponse(c, fiber.StatusOK, "Defecta item updated successfully", defectaItem)
}

// DeleteDefectaItem menangani penghapusan item defecta yang ada.
func DeleteDefectaItem(c *fiber.Ctx) error {
	db := configs.DB
	id := c.Params("id")

	// Cek apakah defecta item dengan ID tersebut ada
	var defectaItem models.DefectaItems
	if err := db.First(&defectaItem, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Defecta item not found", nil)
	}

	// Hapus defecta item dari database
	if err := db.Delete(&defectaItem).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to delete defecta item", nil)
	}

	// Update total_estimate
	var totalEstimate int
	db.Table("defecta_items").Where("defecta_id = ?", defectaItem.DefectaId).Select("COALESCE(SUM(sub_total), 0)").Row().Scan(&totalEstimate)
	db.Table("defectas").Where("id = ?", defectaItem.DefectaId).Update("total_estimate", totalEstimate)

	// Kembalikan respons sukses
	return helpers.JSONResponse(c, fiber.StatusOK, "Defecta item deleted successfully", nil)
}

func GetAllDefectas(c *fiber.Ctx) error {
	// Dapatkan waktu sekarang di WIB
	nowWIB := time.Now().In(configs.Location)

	// Ambil informasi dari token melalui middleware
	branchID, _ := services.GetBranchID(c)

	// Ambil parameter query dan search dari query URL
	pageParam := c.Query("page")
	search := strings.TrimSpace(c.Query("search"))

	page := 1
	if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
		page = p
	}

	limit := 10                  // Tetapkan limit ke 10 data per halaman
	offset := (page - 1) * limit // Hitung offset berdasarkan halaman dan limit

	month := strings.TrimSpace(c.Query("month"))

	// Jika month kosong, isi dengan bulan ini (format YYYY-MM)
	if month == "" {
		month = nowWIB.Format("2006-01")
	}

	// Inisialisasi slice untuk menampung defectas dan variabel total
	var defectas []models.Defectas
	var total int64

	// Bangun query dasar
	query := configs.DB.Table("defectas df").
		Select("df.id, df.defecta_date, df.total_estimate, df.defecta_status").
		Where("df.branch_id = ?", branchID)

	// Filter berdasarkan bulan
	startDate, err := time.Parse("2006-01", month)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid month format. Use YYYY-MM", nil)
	}
	endDate := startDate.AddDate(0, 1, 0)
	query = query.Where("df.defecta_date >= ? AND df.defecta_date < ?", startDate, endDate)

	// Terapkan pencarian jika ada
	if search != "" {
		likeSearch := "%" + search + "%"
		query = query.Where("df.id LIKE ? OR df.defecta_status LIKE ?", likeSearch, likeSearch)
	}

	// Hitung total data untuk pagination
	if err := query.Count(&total).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to count defectas", nil)
	}

	// Ambil data dengan pagination
	if err := query.Order("df.created_at DESC").Limit(limit).Offset(offset).Find(&defectas).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to fetch defectas", nil)
	}

	// Hitung total halaman
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// Format data defectas sebelum dikirimkan dalam respons
	var formattedDefectas []models.DefectaDetailResponse
	for _, d := range defectas {
		formattedDefectas = append(formattedDefectas, models.DefectaDetailResponse{
			ID:            d.ID,
			DefectaDate:   helpers.FormatIndonesianDate(d.DefectaDate),
			TotalEstimate: d.TotalEstimate,
			DefectaStatus: string(d.DefectaStatus),
		})
	}

	// Siapkan data respons dengan pagination
	return helpers.JSONResponseGetAll(c, fiber.StatusOK, "Defectas retrieved successfully", search, int(total), page, totalPages, limit, formattedDefectas)
}

// GetAllDefectaItems menangani pengambilan semua item defecta untuk defecta tertentu.
func GetAllDefectaItems(c *fiber.Ctx) error {
	// Ambil parameter defectaID dari URL
	defectaID := c.Params("id")

	// Inisialisasi slice untuk menampung defecta items
	var defectaItems []models.AllDefectaItems

	// Bangun query untuk mengambil defecta items beserta nama produk dan unitnya
	query := configs.DB.Table("defecta_items di").
		Select("di.id, di.defecta_id, pro.name as product_name, un.name as unit_name, di.price, di.qty, di.sub_total").
		Joins("LEFT JOIN products pro ON pro.id = di.product_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("di.defecta_id = ?", defectaID)

	// Eksekusi query
	if err := query.Find(&defectaItems).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to fetch defecta items", nil)
	}

	// Kembalikan respons sukses dengan data defecta items
	return helpers.JSONResponse(c, fiber.StatusOK, "Defecta items retrieved successfully", defectaItems)
}

func GetDefetaWithItems(c *fiber.Ctx) error {
	defectaID := c.Params("id")
	db := configs.DB

	// Ambil data defecta
	var defecta models.Defectas
	if err := db.First(&defecta, "id = ?", defectaID).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "Defecta not found", nil)
	}

	// Ambil data item defecta beserta nama produk dan unitnya
	var defectaItems []models.AllDefectaItems
	if err := db.Table("defecta_items di").
		Select("di.id, di.product_id, pro.name as product_name, di.unit_id, un.name as unit_name, di.price, di.qty, di.sub_total").
		Joins("LEFT JOIN products pro ON pro.id = di.product_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("di.defecta_id = ?", defecta.ID).
		Find(&defectaItems).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to fetch defecta items", nil)
	}

	var formatedDefectaItems []models.AllDefectaItems
	for _, item := range defectaItems {
		formatedDefectaItems = append(formatedDefectaItems, models.AllDefectaItems{
			ID:          item.ID,
			ProductId:   item.ProductId,
			ProductName: item.ProductName,
			UnitId:      item.UnitId,
			UnitName:    item.UnitName,
			Price:       item.Price,
			Qty:         item.Qty,
			SubTotal:    item.SubTotal,
		})
	}

	formatedDefetaDate := helpers.FormatIndonesianDate(defecta.DefectaDate)

	// Siapkan respons dengan detail defecta dan itemnya
	response := models.DefectaDetailWithItemsResponse{
		ID:            defecta.ID,
		DefectaDate:   formatedDefetaDate,
		TotalEstimate: defecta.TotalEstimate,
		DefectaStatus: string(defecta.DefectaStatus),
		Items:         formatedDefectaItems,
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Defecta details retrieved successfully", response)
}
