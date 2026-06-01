package controllers

import (
	context "context"
	json "encoding/json"
	fmt "fmt"
	log "log"
	os "os"
	strings "strings"
	time "time"

	fiber "github.com/gofiber/fiber/v2"
	jwt "github.com/golang-jwt/jwt/v5"
	config "github.com/heru-oktafian/fiber-apotek/configs"
	helpers "github.com/heru-oktafian/fiber-apotek/helpers"
	models "github.com/heru-oktafian/fiber-apotek/models"
	services "github.com/heru-oktafian/fiber-apotek/services"
	bcrypt "golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// generateJWT menghasilkan token JWT untuk pengguna
func generateJWT(user models.User) (string, error) {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(config.Location)

	// Definisikan klaim JWT
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": nowWIB.Add(5 * time.Minute).Unix(),
	}
	// Buat token menggunakan klaim dan kunci penandatanganan
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Ganti dengan kunci penandatanganan Anda yang sebenarnya (misalnya, variabel environment)
	secretKey := []byte(os.Getenv("JWT_SECRET"))
	return token.SignedString(secretKey)
}

// Fungsi untuk menambahkan token ke blacklist di Redis dengan TTL 8 jam
func blacklistToken(token string) error {
	// Parse token untuk mendapatkan waktu kedaluwarsa (exp)
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		secretKey := []byte(os.Getenv("JWT_SECRET"))
		return secretKey, nil
	})

	if err != nil || !parsedToken.Valid {
		log.Printf("Failed to parse token: %v", err)
		return err
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || claims["exp"] == nil {
		log.Println("Invalid token claims, no exp found")
		return fmt.Errorf("invalid token claims")
	}

	// Hitung waktu kedaluwarsa token
	expiryUnix := int64(claims["exp"].(float64)) // Klaim `exp` adalah float64
	expiryTime := time.Unix(expiryUnix, 0)
	ttl := time.Until(expiryTime)

	// Pastikan TTL valid
	if ttl <= 0 {
		log.Println("Token already expired")
		return fmt.Errorf("token already expired")
	}

	// Tambahkan token ke Redis dengan TTL
	ctx := context.Background()
	redisKey := fmt.Sprintf("blacklist:%s", token)
	err = config.RDB.Set(ctx, redisKey, "blacklisted", ttl).Err()
	if err != nil {
		log.Printf("Failed to blacklist token: %v", err)
		return err
	}

	log.Printf("Token blacklisted successfully with TTL: %v", ttl)
	return nil
}

// Function Login menangani login pengguna
func Login(c *fiber.Ctx) error {
	// Definisikan variabel loginRequest dan user
	var loginRequest LoginRequest
	var user models.User

	// Parse input JSON menjadi struct LoginRequest
	if err := c.BodyParser(&loginRequest); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid input", err)
	}

	// Cari user berdasarkan username
	if err := config.DB.Where("username = ? AND user_status = 'active'", loginRequest.Username).First(&user).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusUnauthorized, "Login failed", "User is not active, call admin to activated your account !")
	}

	// Bandingkan password input dengan password yang sudah di-hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
		return helpers.JSONResponse(c, fiber.StatusUnauthorized, "Login failed", "Invalid username or password")
	}

	// Buat token JWT
	token, err := generateJWT(user)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Login failed", "Failed to generate token")
	}

	// Jika username dan password cocok, lanjutkan proses (misalnya buat token JWT)
	return helpers.JSONResponse(c, fiber.StatusOK, "Login successful", token)
}

// Function Logout menangani logout pengguna
func Logout(c *fiber.Ctx) error {
	// Ambil token dari header Authorization
	token := c.Get("Authorization")

	// Remove prefix "Bearer " jika ada
	token = strings.TrimPrefix(token, "Bearer ")

	if token == "" {
		return helpers.JSONResponse(c, fiber.StatusUnauthorized, "Missing token", "Insert valid token to access this endpoint !")
	}

	// Blacklist token JWT
	if err := blacklistToken(token); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Logout failed", "Failed to blacklist token")
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Logout successful", "Logout successful")
}

func SetBranch(c *fiber.Ctx) error {
	// Ambil token dari header Authorization
	token := c.Get("Authorization")

	// Hapus prefix "Bearer " jika ada
	token = strings.TrimPrefix(token, "Bearer ")

	// Periksa jika token kosong
	if token == "" {
		return helpers.JSONResponse(c, fiber.StatusUnauthorized, "Missing token", "Insert valid token to access this endpoint!")
	}

	// Verifikasi token JWT untuk mendapatkan user ID
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		secretKey := []byte(os.Getenv("JWT_SECRET"))
		return secretKey, nil
	})

	if err != nil || !parsedToken.Valid {
		return helpers.JSONResponse(c, fiber.StatusUnauthorized, "Invalid token", "Try to login again!")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || claims["sub"] == nil {
		return helpers.JSONResponse(c, fiber.StatusUnauthorized, "Invalid token claims", "Try to login again!")
	}

	// Ambil user ID dari klaim token
	userID := string(claims["sub"].(string))

	// Parse input JSON untuk mendapatkan branch ID
	var request struct {
		BranchID string `json:"branch_id" validate:"required"`
	}
	if err := c.BodyParser(&request); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid input", err)
	}

	// Periksa apakah branch_id valid untuk user ini
	var userBranch models.UserBranch
	if err := config.DB.Where("user_id = ? AND branch_id = ?", userID, request.BranchID).First(&userBranch).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusForbidden, "Invalid branch ID", "Branch not associated with this user!")
	}

	// Ambil user_role dari tabel users berdasarkan user_id
	var user models.User
	if err := config.DB.Select("name AS name, user_role AS user_role").Where("id = ?", userID).First(&user).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to set branch", "Unable to retrieve user role")
	}

	// Ambil default_member, quota, subscription_type, dan real_asset dari branch
	var branch models.Branch
	if err := config.DB.Select("default_member, quota, subscription_type, real_asset").Where("id = ?", request.BranchID).First(&branch).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to set branch", "Unable to retrieve branch details")
	}

	// Buat token JWT baru dengan klaim branch_id dan user_role
	newToken, err := generateBranchJWTWithRole(userID, request.BranchID, string(user.UserRole), branch.DefaultMember, branch.Quota, string(branch.SubscriptionType), string(branch.RealAsset), user.Name)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to set branch", "Failed to generate new token")
	}

	// Tambahkan token lama ke Redis blacklist
	if err := blacklistToken(token); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to set branch", "Failed to blacklist old token")
	}

	// Berikan token baru ke pengguna
	return helpers.JSONResponse(c, fiber.StatusOK, "Branch set successfully", newToken)
}

func generateBranchJWTWithRole(userID string, branchID string, userRole string, defaultMember string, quota int, subscriptionType string, realAsset string, namaUser string) (string, error) {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(config.Location)

	// Definisikan klaim untuk token baru
	claims := jwt.MapClaims{
		"sub":               userID,                           // User ID
		"name":              namaUser,                         // Nama User
		"branch_id":         branchID,                         // Branch ID
		"user_role":         userRole,                         // User Role
		"exp":               nowWIB.Add(8 * time.Hour).Unix(), // Expired dalam 8 jam
		"default_member":    defaultMember,
		"quota":             quota,
		"subscription_type": subscriptionType,
		"real_asset":        realAsset,
	}

	// Buat token baru dengan klaim
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Gunakan secret key untuk menandatangani token
	secretKey := []byte(os.Getenv("JWT_SECRET"))
	return token.SignedString(secretKey)
}

// GetProfile menangani penampilan branch sesuai branch_id dari tokenJWT
func GetProfile(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	userID, _ := services.GetUserID(c)
	userRole, _ := services.GetUserRole(c)
	var profilStruct models.ProfileStruct

	// Melakukan LEFT OUTER JOIN menggunakan GORM
	if err := config.DB.
		Table("user_branches usrbrc").
		Select("usrbrc.user_id AS user_id, usr.name AS profile_name, usrbrc.branch_id AS branch_id, brc.branch_name AS branch_name, brc.address, brc.phone, brc.email, brc.sia_id, brc.sia_name, brc.psa_id, brc.psa_name, brc.sipa, brc.sipa_name, brc.aping_id, brc.aping_name, brc.bank_name, brc.account_name, brc.account_number, brc.tax_percentage, brc.journal_method, brc.branch_status, brc.license_date, brc.default_member AS default_member, mbr.name AS member_name, brc.real_asset AS real_asset").
		Joins("LEFT JOIN users usr ON usr.id = usrbrc.user_id").
		Joins("LEFT JOIN branches brc ON brc.id = usrbrc.branch_id").
		Joins("LEFT JOIN members mbr ON mbr.id = brc.default_member").
		Where("usrbrc.branch_id = ? AND usrbrc.user_id = ?", branchID, userID).
		Scan(&profilStruct).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get userbranches failed", "Failed to fetch user branches with details")
	}

	// Mengembalikan response data branch
	return helpers.JSONResponse(c, fiber.StatusOK, "Otoritas : "+userRole, profilStruct)
}

// Coba
func Coba(c *fiber.Ctx) error {
	return c.SendString("Coba berhasil!")
}

// GetMenus handles the request to retrieve menu data with an optional user_role filter.
func GetMenus(c *fiber.Ctx) error {
	userRoles, _ := services.GetUserRole(c)

	// Read the menus.json file
	data, err := os.ReadFile("menus.json")
	if err != nil {
		log.Printf("Error reading menus.json: %v", err)
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to read menu data", err)
	}

	var menuResponse models.MenuResponse // Gunakan struct pembungkus baru
	err = json.Unmarshal(data, &menuResponse)
	if err != nil {
		log.Printf("Error unmarshaling menus.json: %v", err)
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to parse menu data", err)
	}

	// Data menu yang sebenarnya sekarang ada di menuResponse.Data
	menus := menuResponse.Data

	// Get the user_role query parameter
	userRoleFilter := userRoles

	if userRoleFilter == "" {
		// If no user_role filter is provided, return all menus
		return helpers.JSONResponse(c, fiber.StatusOK, "Get All Menus Success", menus)
	}

	// Filter menus by user_role
	var filteredMenus []models.Menu // Gunakan models.Menu
	for _, menu := range menus {
		if strings.EqualFold(menu.UserRole, userRoleFilter) {
			filteredMenus = append(filteredMenus, menu)
		}
	}

	if len(filteredMenus) == 0 {
		return helpers.JSONResponse(c, fiber.StatusNotFound, "No menu found for the specified user_role", []models.Menu{})
	}

	return helpers.JSONResponse(c, fiber.StatusOK, "Get Menus by User Role Success", filteredMenus)
}
