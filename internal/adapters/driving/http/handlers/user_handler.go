package handlers

import (
	strconv "strconv"
	strings "strings"

	configs "apotek-clean/configs"
	helpers "apotek-clean/helpers"
	models "apotek-clean/models"
	services "apotek-clean/services"
	fiber "github.com/gofiber/fiber/v2"
	bcrypt "golang.org/x/crypto/bcrypt"
	gorm "gorm.io/gorm"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	pageParam := c.Query("page")
	search := strings.TrimSpace(c.Query("search"))
	page := 1
	if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
		page = p
	}
	limit := 10
	offset := (page - 1) * limit

	var users []models.User
	db := configs.DB.Model(&models.User{}).Omit("Password")
	if search != "" {
		searchPattern := "%" + search + "%"
		db = db.Where("username ILIKE ? OR name ILIKE ? ", searchPattern, searchPattern)
	}

	var total int64
	db.Count(&total)
	if err := db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data user", err)
	}
	for i := range users {
		users[i].Password = ""
	}
	return helpers.JSONResponseFlat(c, fiber.StatusOK, "Data berhasil diambil", map[string]interface{}{
		"per_page":     limit,
		"current_page": page,
		"search":       search,
		"total":        int(total),
		"total_pages":  int((total + int64(limit) - 1) / int64(limit)),
		"data":         users,
	})
}

func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	user, err := services.FindUserByID(configs.DB.Omit("Password"), userID)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "User tidak ditemukan", err)
	}
	user.Password = ""
	return helpers.JSONResponse(c, fiber.StatusOK, "Data user berhasil ditemukan", user)
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	user := new(models.User)
	if err := c.BodyParser(user); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Lengkapi data user yang ingin dibuat", err)
	}
	if user.Username == "" || user.UserRole == "" || user.Name == "" {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Username, Password, Name dan Role harus diisi", nil)
	}
	allowedRoles := map[string]bool{
		"administrator": true, "superadmin": true, "operator": true, "cashier": true,
		"finance": true, "pendaftaran": true, "rekammedis": true, "ralan": true,
		"ranap": true, "vk": true, "lab": true, "klaim": true, "simrs": true,
		"ipsrs": true, "umum": true,
	}
	if !allowedRoles[string(user.UserRole)] {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid user role: "+string(user.UserRole), nil)
	}
	if user.UserStatus == "" {
		user.UserStatus = "inactive"
	} else {
		allowedStatuses := map[string]bool{"active": true, "inactive": true}
		if !allowedStatuses[string(user.UserStatus)] {
			return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid user status: "+string(user.UserStatus), nil)
		}
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Could not hash password", err)
	}
	user.Password = string(hashedPassword)
	user.ID = helpers.GenerateID("USR")
	result := configs.DB.Create(&user)
	if result.Error != nil {
		if helpers.IsDuplicateKeyError(result.Error) {
			return helpers.JSONResponse(c, fiber.StatusBadRequest, "Username sudah digunakan", result.Error)
		}
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal membuat user", result.Error)
	}
	configs.RDB.Del(configs.Ctx, "/api/users")
	user.Password = ""
	return helpers.JSONResponse(c, fiber.StatusCreated, "User berhasil dibuat", user)
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	user, err := services.FindUserByID(configs.DB, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return helpers.JSONResponse(c, fiber.StatusNotFound, "User tidak ditemukan", nil)
		}
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal menemukan user untuk update", err)
	}
	updateData := new(struct {
		Username   string `json:"username"`
		Name       string `json:"name"`
		Password   string `json:"password"`
		UserRole   string `json:"user_role"`
		UserStatus string `json:"user_status"`
	})
	if err := c.BodyParser(updateData); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Format data yang dikirim tidak valid", err)
	}
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
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updateData.Password), bcrypt.DefaultCost)
		if err != nil {
			return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Could not hash new password", err)
		}
		user.Password = string(hashedPassword)
	}
	result := configs.DB.Save(user)
	if result.Error != nil {
		if helpers.IsDuplicateKeyError(result.Error) {
			return helpers.JSONResponse(c, fiber.StatusBadRequest, "Username sudah digunakan", result.Error)
		}
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Gagal mengupdate user", result.Error)
	}
	configs.RDB.Del(configs.Ctx, "/api/users", "/api/users/"+userID)
	user.Password = ""
	return helpers.JSONResponse(c, fiber.StatusOK, "User berhasil diupdate", user)
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	userID := c.Params("user_id")
	user, err := services.FindUserByID(configs.DB, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return helpers.JSONResponse(c, fiber.StatusNotFound, "User not found", nil)
		}
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to retrieve user for deletion", err)
	}
	result := configs.DB.Delete(user)
	if result.Error != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to delete user", result.Error)
	}
	configs.RDB.Del(configs.Ctx, "/api/users", "/api/users/"+userID)
	return helpers.JSONResponse(c, fiber.StatusOK, "User deleted successfully", nil)
}
