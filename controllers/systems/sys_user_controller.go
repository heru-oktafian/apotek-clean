package controllers

import (
	strconv "strconv"
	strings "strings"

	fiber "github.com/gofiber/fiber/v2"
	configs "github.com/heru-oktafian/fiber-apotek/configs"
	helpers "github.com/heru-oktafian/fiber-apotek/helpers"
	models "github.com/heru-oktafian/fiber-apotek/models"
	bcrypt "golang.org/x/crypto/bcrypt"
	gorm "gorm.io/gorm"
)

// GetUsers mengambil semua pengguna dengan paginasi dan pencarian.
func GetUsers(c *fiber.Ctx) error {

	// Ambil parameter page dan search dari query URL
	pageParam := c.Query("page")
	search := strings.TrimSpace(c.Query("search"))

	// Konversi page ke int, default ke 1 jika tidak valid
	page := 1
	if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
		page = p
	}

	limit := 10                  // Tetapkan batas data per halaman ke 10
	offset := (page - 1) * limit // Hitung offset berdasarkan halaman dan limit

	var users []models.User
	db := configs.DB.Model(&models.User{}).Omit("Password")

	// Pencarian berdasarkan username, name, atau user_role
	if search != "" {
		searchPattern := "%" + search + "%"
		db = db.Where("username ILIKE ? OR name ILIKE ? ", searchPattern, searchPattern)
	}

	var total int64
	db.Count(&total)

	// Ambil data dengan paginasi
	if err := db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data user", err)
	}

	// Hilangkan password dari hasil response
	for i := range users {
		users[i].Password = ""
	}

	// Kembalikan hasil response tanpa nested "data"
	return helpers.JSONResponseFlat(c, fiber.StatusOK, "Data berhasil diambil", map[string]interface{}{
		"per_page":     limit,
		"current_page": page,
		"search":       search,
		"total":        int(total),
		"total_pages":  int((total + int64(limit) - 1) / int64(limit)),
		"data":         users,
	})
}

// GetUserByID mengambil pengguna berdasarkan USER_Id.
func GetUserByID(c *fiber.Ctx) error {
	// Ambil USER_Id dari parameter URL
	UserID := c.Params("user_id")

	// Buat instance kosong dari model User untuk menampung data
	var user models.User

	// Panggil helper GetResource.
	// Kita perlu menambahkan Omit("Password") dan filter Where("user_id = ?", UserID)
	// sebelum melewatkan DB instance ke GetResource.
	// GetResource akan melanjutkan dengan First(&user)
	dbQuery := configs.DB.Omit("Password").Where("user_id = ?", UserID)

	err := helpers.GetResource(c, dbQuery, &user, UserID) // UserID di sini hanya sebagai placeholder untuk `id` di GetResource
	if err != nil {
		// GetResource sudah menangani respons Not Found dan Internal Server Error,
		// jadi kita hanya perlu mengembalikan error tersebut.
		return err
	}

	// GetResource juga sudah mengurus pengiriman respons JSON untuk data yang berhasil ditemukan.
	return nil
}

// CreateUser membuat pengguna baru. Hanya 'administrator' & 'superadmin'
func CreateUser(c *fiber.Ctx) error {
	user := new(models.User)
	if err := c.BodyParser(user); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Lengkapi data user yang ingin dibuat", err)
	}

	// Basic validasi input mandatory
	if user.Username == "" || user.UserRole == "" || user.Name == "" {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Username, Password, Name dan Role harus diisi", nil)
	}

	// Validate user_role against allowed ENUM values
	allowedRoles := map[string]bool{
		"administrator": true, "superadmin": true, "operator": true, "cashier": true,
		"finance": true, "pendaftaran": true, "rekammedis": true, "ralan": true,
		"ranap": true, "vk": true, "lab": true, "klaim": true, "simrs": true,
		"ipsrs": true, "umum": true,
	}
	if !allowedRoles[string(user.UserRole)] {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid user role: "+string(user.UserRole), nil)
	}

	// Validate user_status (optional, default will be 'inactive' by GORM)
	if user.UserStatus == "" {
		user.UserStatus = "inactive" // Default if not provided
	} else {
		allowedStatuses := map[string]bool{"active": true, "inactive": true}
		if !allowedStatuses[string(user.UserStatus)] {
			return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid user status: "+string(user.UserStatus), nil)
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		// log.Printf("Error hashing password during user creation: %v", err)
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Could not hash password", err)
	}
	user.Password = string(hashedPassword)

	// Generate custom USER_Id
	user.ID = helpers.GenerateID("USR")

	result := configs.DB.Create(&user)
	if result.Error != nil {
		if helpers.IsDuplicateKeyError(result.Error) {
			return helpers.JSONResponse(c, fiber.StatusBadRequest, "Username sudah digunakan", result.Error)
		}
		// log.Printf("Error creating user: %v", result.Error)
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal membuat user", result.Error)
	}

	// Invalidate relevant cache (e.g., list of users)
	configs.RDB.Del(configs.Ctx, "/api/users")

	// Return user without password
	user.Password = ""
	return helpers.JSONResponse(c, fiber.StatusCreated, "User berhasil dibuat", user)
}

// UpdateUser memperbarui pengguna berdasarkan USER_Id. Hanya 'administrator' & 'superadmin'
func UpdateUser(c *fiber.Ctx) error {
	UserID := c.Params("user_id")
	var user models.User
	result := configs.DB.Where("user_id = ?", UserID).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return helpers.JSONResponse(c, fiber.StatusNotFound, "User tidak ditemukan", nil)
		}
		// log.Printf("Error finding user for update: %v", result.Error)
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal menemukan user untuk update", result.Error)
	}

	// Bind request body to a temporary struct to handle optional fields and password change
	updateData := new(struct {
		Username   string `json:"username"`
		Name       string `json:"name"`
		Password   string `json:"password"` // Optional: new password
		UserRole   string `json:"user_role"`
		UserStatus string `json:"user_status"`
	})
	if err := c.BodyParser(updateData); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Format data yang dikirim tidak valid", err)
	}

	// Update fields if provided
	if updateData.Username != "" {
		user.Username = updateData.Username
	}

	if updateData.Name != "" {
		user.Name = updateData.Name
	}

	if updateData.UserRole != "" {
		user.UserRole = models.UserRole(updateData.UserRole)
	}

	if updateData.UserStatus != "" {
		user.UserStatus = models.DataStatus(updateData.UserStatus)
	}

	if updateData.Password != "" {
		// Hash new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updateData.Password), bcrypt.DefaultCost)
		if err != nil {
			// log.Printf("Error hashing new password during user update: %v", err)
			return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Could not hash new password", err)
		}
		user.Password = string(hashedPassword)
	}

	result = configs.DB.Save(&user)
	if result.Error != nil {
		if helpers.IsDuplicateKeyError(result.Error) {
			return helpers.JSONResponse(c, fiber.StatusBadRequest, "Username sudah digunakan", result.Error)
		}
		// log.Printf("Error updating user: %v", result.Error)
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengupdate user", result.Error)
	}

	// Invalidate relevant cache
	configs.RDB.Del(configs.Ctx, "/api/users", "/api/users/"+UserID)

	// Return updated user without password
	user.Password = ""
	return helpers.JSONResponse(c, fiber.StatusOK, "User berhasil diupdate", user)
}

// DeleteUser menghapus pengguna berdasarkan USER_ID (soft delete). Hanya 'administrator' & 'superadmin'
func DeleteUser(c *fiber.Ctx) error {
	UserID := c.Params("user_id")
	var user models.User
	result := configs.DB.Where("user_id = ?", UserID).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return helpers.JSONResponse(c, fiber.StatusNotFound, "User not found", nil)
		}
		// log.Printf("Error finding user for deletion: %v", result.Error)
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to retrieve user for deletion", result.Error)
	}

	// Lakukan soft delete (GORM akan mengisi kolom DeletedAt)
	result = configs.DB.Delete(&user)
	if result.Error != nil {
		// log.Printf("Error deleting user: %v", result.Error)
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to delete user", result.Error)
	}

	// Invalidate relevant cache
	configs.RDB.Del(configs.Ctx, "/api/users", "/api/users/"+UserID)

	return helpers.JSONResponse(c, fiber.StatusOK, "User deleted successfully", nil)
}
