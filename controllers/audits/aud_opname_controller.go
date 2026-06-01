package controllers

import (
	fmt "fmt"
	math "math"
	http "net/http"
	strconv "strconv"
	strings "strings"
	time "time"

	fiber "github.com/gofiber/fiber/v2"
	configs "github.com/heru-oktafian/fiber-apotek/configs"
	helpers "github.com/heru-oktafian/fiber-apotek/helpers"
	models "github.com/heru-oktafian/fiber-apotek/models"
	services "github.com/heru-oktafian/fiber-apotek/services"
	gorm "gorm.io/gorm"
)

// Get Mobile Opnames menampilkan semua opname yang disajikan untuk pengguna mobile
func GetAllMobileOpnames(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	var rawOpnames []models.OpnameQueryResult // Gunakan struct untuk menampung hasil query mentah

	// Query dasar (opname_date tanpa TO_CHAR di SQL)
	query := configs.DB.Table("opnames pur").
		Select("pur.id, pur.description, pur.opname_date, 'Rp. ' || TO_CHAR(pur.total_opname, 'FM999G999G999') AS total_opname").
		Where("pur.branch_id = ?", branchID).
		Order("pur.created_at DESC")

	if err := query.Scan(&rawOpnames).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Pengambilan opnames gagal", "Gagal mengambil data Opname")
	}

	// Inisialisasi slice untuk hasil akhir yang diformat
	var formattedOpnames []models.AllOpnameMobiles

	// Iterasi hasil query mentah dan format opname_date
	for _, op := range rawOpnames {
		formattedOpnames = append(formattedOpnames, models.AllOpnameMobiles{
			ID:          op.ID,
			Description: op.Description,
			OpnameDate:  helpers.FormatIndonesianDate(op.OpnameDate),
			TotalOpname: op.TotalOpname,
		})
	}

	return helpers.JSONResponse(c, http.StatusOK, "Data opname berhasil diambil", formattedOpnames)
}

// Get Mobile Opnames Actives menampilkan semua opname yang disajikan untuk pengguna mobile dengan status aktif
func GetAllActiveMobileOpnames(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	var rawOpnames []models.OpnameQueryResult // Gunakan struct untuk menampung hasil query mentah

	// Query dasar (opname_date tanpa TO_CHAR di SQL)
	query := configs.DB.Table("opnames pur").
		Select("pur.id, pur.description, pur.opname_date, 'Rp. ' || TO_CHAR(pur.total_opname, 'FM999G999G999') AS total_opname").
		Where("pur.branch_id = ? AND pur.opname_status = 'active' ", branchID).
		Order("pur.created_at DESC")

	if err := query.Scan(&rawOpnames).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Pengambilan Opname gagal", "Gagal mengambil data Opname")
	}

	// Inisialisasi slice untuk hasil akhir yang diformat
	var formattedOpnames []models.AllOpnameMobiles

	// Iterasi hasil query mentah dan format opname_date
	for _, op := range rawOpnames {
		formattedOpnames = append(formattedOpnames, models.AllOpnameMobiles{
			ID:          op.ID,
			Description: op.Description,
			OpnameDate:  helpers.FormatIndonesianDate(op.OpnameDate), // <--- Gunakan helper di sini!
			TotalOpname: op.TotalOpname,
		})
	}

	return helpers.JSONResponse(c, http.StatusOK, "Data opname berhasil diambil", formattedOpnames)
}

// GetMobileOpnameItemDetails adalah fungsi menammpilkan semua item berdasarkan product_name tanpa pagination
func GetMobileOpnameItemDetails(c *fiber.Ctx) error {
	// Get branch id
	branchID, _ := services.GetBranchID(c)

	// Parsing body JSON ke struct
	var OpnameItems []models.AllOpnameItemDetails

	// Query dasar
	query := configs.DB.Table("opname_items pit").
		Select("pit.id, pit.opname_id, pit.product_id, pro.name AS product_name, pit.price, (pit.qty - pit.qty_exist) AS qty_adjustment, (pit.sub_total - pit.sub_total_exist) AS sub_adjustment, pit.expired_date").
		Joins("LEFT JOIN products pro ON pro.id = pit.product_id").
		Joins("LEFT JOIN opnames opn ON opn.id = pit.opname_id").
		Where("opn.branch_id = ?", branchID).
		Order("pro.name ASC")

	// Eksekusi query
	if err := query.Scan(&OpnameItems).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Pengambilan item gagal", "Gagal mengambil data item Opname")
	}

	return helpers.JSONResponse(c, http.StatusOK, "Item berhasil ditampilkan", OpnameItems)
}

// GetMobileOpnameItemsGlimpse adalah fungsi menammpilkan 5 item berdasarkan product_name tanpa pagination
func GetMobileOpnameItemsGlimpse(c *fiber.Ctx) error {
	// Get branch id
	branchID, _ := services.GetBranchID(c)

	// Parsing body JSON ke struct
	var OpnameItems []models.AllOpnameItemDetails

	// Query dasar
	query := configs.DB.Table("opname_items pit").
		Select("pit.id, pit.opname_id, pit.product_id, pro.name AS product_name, pit.price, (pit.qty - pit.qty_exist) AS qty_adjustment, (pit.sub_total - pit.sub_total_exist) AS sub_adjustment, pit.expired_date").
		Joins("LEFT JOIN products pro ON pro.id = pit.product_id").
		Joins("LEFT JOIN opnames opn ON opn.id = pit.opname_id").
		Where("opn.branch_id = ?", branchID).
		Limit(5)

	// Eksekusi query
	if err := query.Scan(&OpnameItems).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Pengambilan item gagal", "Gagal mengambil data item Opname")
	}

	return helpers.JSONResponse(c, http.StatusOK, "Item berhasil ditampilkan", OpnameItems)
}

// Get All Opnames tampilkan semua opname
func GetAllOpnames(c *fiber.Ctx) error {
	// Dapatkan waktu sekarang di WIB
	nowWIB := time.Now().In(configs.Location)

	// Get branch id
	branch_id, _ := services.GetBranchID(c)

	// Ambil parameter page dan search dari query URL
	pageParam := c.Query("page")
	search := strings.TrimSpace(c.Query("search"))
	month := strings.TrimSpace(c.Query("month"))

	// Jika month kosong, isi dengan bulan ini (format YYYY-MM)
	if month == "" {
		month = nowWIB.Format("2006-01")
	}

	startDate, err := time.Parse("2006-01", month)
	if err != nil {
		return helpers.JSONResponse(c, http.StatusBadRequest, "Invalid month format. Use YYYY-MM", nil)
	}
	endDate := startDate.AddDate(0, 1, 0)

	// Konversi page ke int, default ke 1 jika tidak valid
	page := 1
	if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
		page = p
	}

	limit := 10 // Tetapkan limit ke 10 data per halaman
	offset := (page - 1) * limit

	var opnames []models.AllOpnames
	var total int64

	// Query dasar
	query := configs.DB.Table("opnames pur").
		Select("pur.id, pur.description, TO_CHAR(pur.opname_date, 'DD-MM-YYYY') AS opname_date, pur.total_opname").
		Where("pur.branch_id = ?", branch_id).
		Where("pur.opname_date >= ? AND pur.opname_date < ?", startDate, endDate).
		Order("pur.created_at DESC")

	// Jika ada search key, tambahkan filter WHERE
	if search != "" {
		search = strings.ToLower(search)
		query = query.Where("LOWER(pur.description) LIKE ?", "%"+search+"%")
	}

	// Hitung total opname yang sesuai dengan filter
	if err := query.Count(&total).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Pengambilan opname gagal", "Gagal menghitung stok awal")
	}

	// Ambil data dengan pagination
	if err := query.Offset(offset).Limit(limit).Scan(&opnames).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Pengambilan opname gagal", "Gagal mengambil data stok awal")
	}

	// Hitung total halaman berdasarkan hasil filter
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// Kembalikan hasil response tanpa nested "data"
	return helpers.JSONResponseFlat(c, http.StatusOK, "Data Opname berhasil diambil", map[string]interface{}{
		"per_page":     limit,
		"current_page": page,
		"search":       search,
		"total":        total,
		"total_pages":  totalPages,
		"data":         opnames,
	})
}

// CreateOpname Function
func CreateOpname(c *fiber.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	db := configs.DB

	// Ambil informasi dari token
	branchID, _ := services.GetBranchID(c)
	userID, _ := services.GetUserID(c)
	generatedID := helpers.GenerateID("OPN")

	// Ambil input dari body
	var input models.OpnameInput
	if err := c.BodyParser(&input); err != nil {
		return helpers.JSONResponse(c, http.StatusBadRequest, "Input tidak valid", err)
	}

	// Parse tanggal
	layout := "2006-01-02" // format harus YYYY-MM-DD
	parsedDate, err := time.Parse(layout, input.OpnameDate)
	if err != nil {
		return helpers.JSONResponse(c, http.StatusBadRequest, "Format tanggal tidak valid. Gunakan YYYY-MM-DD", err)
	}

	// Map ke struct model
	opname := models.Opnames{
		ID:          generatedID,
		Description: input.Description,
		BranchID:    branchID,
		UserID:      userID,
		OpnameDate:  parsedDate,
		TotalOpname: 0,
		CreatedAt:   nowWIB,
		UpdatedAt:   nowWIB,
	}

	// Simpan opname
	if err := db.Create(&opname).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal menyimpan data opname", err.Error())
	}

	// Buat laporan
	if err := helpers.SyncOpnameReport(db, opname); err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal membuat laporan opname", err.Error())
	}

	_ = helpers.AutoCleanupOpnames(db)

	return helpers.JSONResponse(c, http.StatusOK, "Opname berhasil dibuat", opname)
}

// UpdateOpnameByID Function
func UpdateOpnameByID(c *fiber.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(configs.Location)

	db := configs.DB
	id := c.Params("id")

	// Cari data opname lama
	var opname models.Opnames
	if err := db.First(&opname, "id = ?", id).Error; err != nil {
		// return c.Status(404).JSON(fiber.Map{"error": "Opname tidak ditemukan"})
		return helpers.JSONResponse(c, http.StatusNotFound, "Opname tidak ditemukan", err)
	}

	// Gunakan struct input
	var input models.OpnameInput
	if err := c.BodyParser(&input); err != nil {
		return helpers.JSONResponse(c, http.StatusBadRequest, "Input tidak valid", err)
	}

	// Cek dan update OpnameDate
	if input.OpnameDate != "" {
		layout := "2006-01-02"
		parsedDate, err := time.Parse(layout, input.OpnameDate)
		if err != nil {
			return helpers.JSONResponse(c, http.StatusBadRequest, "Format tanggal tidak valid. Gunakan YYYY-MM-DD", err)
		}
		opname.OpnameDate = parsedDate
	}

	// Cek dan update Payment
	if input.Payment != "" {
		opname.Payment = models.PaymentStatus(input.Payment)
	}

	opname.UpdatedAt = nowWIB

	// Cek dan update Description
	if input.Description != "" {
		opname.Description = input.Description
	}

	// Simpan perubahan dasar terlebih dahulu
	if err := db.Save(&opname).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal memperbarui opname", err)
	}

	// Hitung ulang total secara sinkron dan simpan hasil database langsung menggunakan helper RecalculateTotalOpname
	// Helper RecalculateTotalOpname juga otomatis melakukan SyncOpnameReport.
	if err := helpers.RecalculateTotalOpname(db, id); err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal menghitung ulang total opname", err)
	}

	_ = helpers.AutoCleanupOpnames(db)

	// Fetch again to return updated object
	db.First(&opname, "id = ?", id)
	return helpers.JSONResponse(c, http.StatusOK, "Opname berhasil diperbarui", opname)
}

// DeleteOpnameByID Function
func DeleteOpnameByID(c *fiber.Ctx) error {
	db := configs.DB
	id := c.Params("id")

	// Ambil opname
	var opname models.Opnames
	if err := db.First(&opname, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusNotFound, "Opname tidak ditemukan", err)
	}

	// Ambil item-item dan rollback stok
	var items []models.OpnameItems
	if err := db.Where("opname_id = ?", id).Find(&items).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal mengambil item opname", err)
	}

	for _, item := range items {
		// Kosongkan stok ke produk asynchronously
		services.ZeroProductStockAsync(db, item.ProductId, item.Qty)
	}

	// Hapus semua item dari opname
	if err := db.Where("opname_id = ?", id).Delete(&models.OpnameItems{}).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal menghapus item opname", err)
	}

	// Hapus laporan transaksi terkait
	if err := db.Where("id = ? AND transaction_type = ?", opname.ID, models.Opname).Delete(&models.TransactionReports{}).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal menghapus laporan transaksi", err)
	}

	// Hapus opname
	if err := db.Delete(&opname).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal menghapus opname", err)
	}

	return helpers.JSONResponse(c, http.StatusOK, "Opname berhasil dihapus", opname)
}

// GetOpnameWithItems menampilkan satu opname beserta semua item-nya
func GetOpnameWithItems(c *fiber.Ctx) error {
	db := configs.DB

	// Ambil ID pembelian dari parameter URL
	opnameID := c.Params("id")

	// Struct untuk data utama opname
	var opname models.AllOpnames

	// Ambil data opname
	err := db.Table("opnames pur").
		Select("pur.id, pur.description, TO_CHAR(pur.opname_date, 'DD-MM-YYYY') AS opname_date, pur.total_opname, pur.payment").
		Where("pur.id = ?", opnameID).
		Scan(&opname).Error

	if err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal mendapatkan opname", err)
	}

	// Ambil item pembelian terkait
	var items []models.AllOpnameItemMobiles
	err = db.Table("opname_items pit").
		Select("pit.id, pit.opname_id, pit.product_id, pro.name AS product_name, pit.price, pit.qty, pit.sub_total, TO_CHAR(pit.expired_date, 'DD-MM-YYYY') AS expired_date").
		Joins("LEFT JOIN products pro ON pro.id = pit.product_id").
		Where("pit.opname_id = ?", opnameID).
		Order("pro.name ASC").
		Scan(&items).Error

	if err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal mendapatkan item Opname", err)
	}

	return helpers.JSONResponse(c, http.StatusOK, "Opname berhasil diambil", map[string]interface{}{
		"id":           opnameID,
		"description":  opname.Description,
		"opname_date":  opname.OpnameDate,
		"total_opname": opname.TotalOpname,
		"items":        items,
	})
}

// CreateOpnameItem Function
// expects JSON body with opname_id, product_id, qty, expired_date
func CreateOpnameItem(c *fiber.Ctx) error {
	db := configs.DB
	var input models.CreateOpnameItemInput

	if err := c.BodyParser(&input); err != nil {
		return helpers.JSONResponse(c, http.StatusBadRequest, "Masukan tidak valid: "+err.Error(), err)
	}
	if input.OpnameId == "" {
		return helpers.JSONResponse(c, http.StatusBadRequest, "Opname ID tidak boleh kosong", nil)
	}

	// opname id already provided in input.OpnameId

	// Ambil data produk untuk mendapatkan price, stock, dan purchase_price
	var product models.Product
	if err := db.Where("id = ?", input.ProductId).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return helpers.JSONResponse(c, http.StatusNotFound, "Produk tidak ditemukan", err)
		}
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal mengambil data produk: "+err.Error(), err)
	}

	layout := "2006-01-02"
	parsedDate, err := time.Parse(layout, input.ExpiredDate)
	if err != nil {
		return helpers.JSONResponse(c, http.StatusBadRequest, "Format tanggal tidak valid. Gunakan YYYY-MM-DD", err)
	}

	// PENTING: Simpan stock LAMA sebelum update
	oldStock := product.Stock
	oldPurchasePrice := product.PurchasePrice

	// Update expired date, stock, dan purchase price produk sesuai inputan
	if err := db.Model(&product).Updates(map[string]interface{}{
		"expired_date":   parsedDate,
		"stock":          input.Qty,
		"purchase_price": input.Price, // tambahkan pembaruan harga beli
	}).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal memperbarui produk: "+err.Error(), err)
	}

	var opnameItem models.OpnameItems
	opnameItem.OpnameId = input.OpnameId
	opnameItem.ProductId = input.ProductId
	opnameItem.Qty = input.Qty
	opnameItem.ExpiredDate = parsedDate
	opnameItem.Price = input.Price
	opnameItem.QtyExist = oldStock                         // Gunakan stock LAMA
	opnameItem.SubTotalExist = oldStock * oldPurchasePrice // Gunakan stock LAMA
	opnameItem.SubTotal = opnameItem.Qty * opnameItem.Price

	var existingItem models.OpnameItems
	err = db.Where("opname_id = ? AND product_id = ?", opnameItem.OpnameId, opnameItem.ProductId).First(&existingItem).Error

	if err == nil {
		existingItem.Qty = opnameItem.Qty
		existingItem.SubTotal = opnameItem.SubTotal
		existingItem.ExpiredDate = opnameItem.ExpiredDate

		if err := db.Save(&existingItem).Error; err != nil {
			// return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal memperbarui item opname: " + err.Error()})
			return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal memperbarui item opname: "+err.Error(), err)
		}

		// Update stok asynchronously
		helpers.OpnameProductStockAsync(db, opnameItem.ProductId, opnameItem.Qty)

		if err := helpers.RecalculateTotalOpname(db, opnameItem.OpnameId); err != nil {
			return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal menghitung ulang total opname: "+err.Error(), err)
		}

		return helpers.JSONResponse(c, http.StatusOK, "Item opname berhasil diperbarui", existingItem)

	} else if err != gorm.ErrRecordNotFound {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Terjadi kesalahan di database saat pengecekan: "+err.Error(), err)
	}

	if opnameItem.ID == "" {
		opnameItem.ID = helpers.GenerateID("OPI")
	}

	if err := db.Create(&opnameItem).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal menambahkan item opname: "+err.Error(), err)
	}

	// Update stok asynchronously
	helpers.OpnameProductStockAsync(db, opnameItem.ProductId, opnameItem.Qty)

	if err := helpers.RecalculateTotalOpname(db, opnameItem.OpnameId); err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal menghitung ulang total opname: "+err.Error(), err)
	}

	return helpers.JSONResponse(c, http.StatusOK, "Item opname berhasil disimpan", opnameItem)
}

// GetAllOpnameItems tampilkan semua item berdasarkan product_name tanpa pagination
// Opname ID harus dikirim dalam body JSON sebagai {"opname_id":"..."}
func GetAllOpnameItems(c *fiber.Ctx) error {
	// parse opname_id from body
	var payload struct {
		OpnameId string `json:"opname_id" validate:"required"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return helpers.JSONResponse(c, http.StatusBadRequest, "Masukan tidak valid: "+err.Error(), err)
	}

	// Query dasar
	var OpnameItems []models.AllOpnameItemMobiles
	query := configs.DB.Table("opname_items pit").
		Select("pit.id, pit.opname_id, pit.product_id, pro.name AS product_name, TO_CHAR(pit.price, 'FM999G999G999') AS price, pit.qty, pit.qty_exist, TO_CHAR(pit.sub_total, 'FM999G999G999') AS sub_total, TO_CHAR(pit.sub_total_exist, 'FM999G999G999') AS sub_total_exist, TO_CHAR(pit.expired_date, 'DD-MM-YYYY') AS expired_date").
		Joins("LEFT JOIN products pro ON pro.id = pit.product_id").
		Where("pit.opname_id = ?", payload.OpnameId).
		Order("pro.name ASC")

	// Eksekusi query
	if err := query.Scan(&OpnameItems).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Pengambilan item gagal", "Gagal mengambil data item Opname")
	}

	return helpers.JSONResponse(c, http.StatusOK, "Item berhasil ditampilkan", OpnameItems)
}

// Update OpnameItem
func UpdateOpnameItemByID(c *fiber.Ctx) error {
	db := configs.DB
	// parse id along with update payload from body
	var payload struct {
		ID string `json:"id" validate:"required"`
		models.CreateOpnameItemUpdate
	}
	if err := c.BodyParser(&payload); err != nil {
		return helpers.JSONResponse(c, http.StatusBadRequest, "Masukan tidak valid", nil)
	}
	if payload.ID == "" {
		return helpers.JSONResponse(c, http.StatusBadRequest, "ID item tidak boleh kosong", nil)
	}
	id := payload.ID

	var existingItem models.OpnameItems
	if err := db.First(&existingItem, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusNotFound, "Item tidak ditemukan", nil)
	}

	updatedItem := payload.CreateOpnameItemUpdate

	// Kosongkan stok lama asynchronously
	services.ZeroProductStockAsync(db, existingItem.ProductId, existingItem.Qty)

	// Tambah stok baru asynchronously
	services.AddProductStockAsync(db, updatedItem.ProductId, updatedItem.Qty)

	// Update item
	existingItem.ProductId = updatedItem.ProductId
	existingItem.Qty = updatedItem.Qty
	existingItem.Price = updatedItem.Price
	existingItem.SubTotal = updatedItem.Price * updatedItem.Qty

	layout := "2006-01-02"
	parsedDate, err := time.Parse(layout, updatedItem.ExpiredDate)
	if err != nil {
		return helpers.JSONResponse(c, http.StatusBadRequest, "Format tanggal tidak valid. Gunakan YYYY-MM-DD", err)
	}
	existingItem.ExpiredDate = parsedDate

	if err := db.Save(&existingItem).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal menyimpan item: "+err.Error(), err)
	}

	// Update expired date dan stock produk sesuai inputan
	if err := db.Model(&models.Product{}).Where("id = ?", updatedItem.ProductId).Updates(map[string]interface{}{
		"expired_date": parsedDate,
		"stock":        updatedItem.Qty,
	}).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal memperbarui produk: "+err.Error(), err)
	}

	// Update harga produk jika harga item lebih tinggi
	if err := services.UpdateProductPriceIfHigher(db, updatedItem.ProductId, updatedItem.Price); err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal memperbarui harga produk: "+err.Error(), err)
	}

	go func() {
		if err := helpers.RecalculateTotalOpname(db, existingItem.OpnameId); err != nil {
			fmt.Printf("Failed to recalculate total opname asynchronously: %v\n", err)
		}
	}()

	return helpers.JSONResponse(c, http.StatusOK, "Item berhasil diperbarui", existingItem)
}

// DeleteOpnameItemByID OpnameItem
func DeleteOpnameItemByID(c *fiber.Ctx) error {
	db := configs.DB
	var body struct {
		ID string `json:"id" validate:"required"`
	}
	if err := c.BodyParser(&body); err != nil {
		return helpers.JSONResponse(c, http.StatusBadRequest, "Masukan tidak valid", err)
	}
	id := body.ID

	var item models.OpnameItems
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusNotFound, "Item tidak ditemukan", err)
	}

	// Subtract stok asynchronously
	services.ReduceProductStockAsync(db, item.ProductId, item.Qty)

	// Hapus item
	if err := db.Delete(&item).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusInternalServerError, "Gagal menghapus item: "+err.Error(), err)
	}

	go func(item models.OpnameItems) {
		if err := helpers.RecalculateTotalOpname(db, item.OpnameId); err != nil {
			fmt.Printf("Failed to recalculate total opname asynchronously: %v\n", err)
		}
	}(item)

	return helpers.JSONResponse(c, http.StatusOK, "Item berhasil dihapus", item)
}

// GetBuyProductsCombobox dengan pencarian berdasarkan body
func GetProductsComboboxByName(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	search := strings.TrimSpace(strings.ToLower(c.Query("search")))

	var prodCombo []models.ComboboxProducts
	query := configs.DB.Table("products pro").
		Select("pro.id AS pro_id, pro.name AS pro_name, pro.unit_id, pro.stock, unt.name AS unit_name, pro.purchase_price AS price").
		Joins("LEFT JOIN units unt ON unt.id = pro.unit_id").
		Where("pro.branch_id = ?", branchID)

	if search != "" {
		like := "%" + search + "%"
		query = query.Where("LOWER(pro.name) LIKE ? OR LOWER(pro.id) LIKE ?", like, like)
	}

	query = query.Order("pro.name ASC")

	if err := query.Scan(&prodCombo).Error; err != nil {
		return helpers.JSONResponse(c, http.StatusNotFound, "Combobox tidak ditemukan", err)
	}

	return helpers.JSONResponse(c, http.StatusOK, "Data Combobox ditemukan", prodCombo)
}
