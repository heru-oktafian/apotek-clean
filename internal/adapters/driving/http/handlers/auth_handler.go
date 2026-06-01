package handlers

import (
	context "context"
	json "encoding/json"
	fmt "fmt"
	log "log"
	os "os"
	strings "strings"
	time "time"

	config "apotek-clean/configs"
	helpers "apotek-clean/helpers"
	models "apotek-clean/models"
	services "apotek-clean/services"
	fiber "github.com/gofiber/fiber/v2"
	jwt "github.com/golang-jwt/jwt/v5"
	bcrypt "golang.org/x/crypto/bcrypt"
)

type AuthHandler struct{}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func generateJWT(user models.User) (string, error) {
	nowWIB := time.Now().In(config.Location)
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": nowWIB.Add(5 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secretKey := []byte(os.Getenv("JWT_SECRET"))
	if len(secretKey) == 0 {
		secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))
	}
	return token.SignedString(secretKey)
}

func blacklistToken(token string) error {
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		secretKey := []byte(os.Getenv("JWT_SECRET"))
		if len(secretKey) == 0 {
			secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))
		}
		return secretKey, nil
	})
	if err != nil || !parsedToken.Valid {
		return err
	}
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || claims["exp"] == nil {
		return fmt.Errorf("invalid token claims")
	}
	expiryUnix := int64(claims["exp"].(float64))
	expiryTime := time.Unix(expiryUnix, 0)
	ttl := time.Until(expiryTime)
	if ttl <= 0 {
		return fmt.Errorf("token already expired")
	}
	ctx := context.Background()
	redisKey := fmt.Sprintf("blacklist:%s", token)
	return config.RDB.Set(ctx, redisKey, "blacklisted", ttl).Err()
}

func generateBranchJWTWithRole(userID string, branchID string, userRole string, defaultMember string, quota int, subscriptionType string, realAsset string, namaUser string) (string, error) {
	nowWIB := time.Now().In(config.Location)
	claims := jwt.MapClaims{
		"sub":               userID,
		"name":              namaUser,
		"branch_id":         branchID,
		"user_role":         userRole,
		"exp":               nowWIB.Add(8 * time.Hour).Unix(),
		"default_member":    defaultMember,
		"quota":             quota,
		"subscription_type": subscriptionType,
		"real_asset":        realAsset,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secretKey := []byte(os.Getenv("JWT_SECRET"))
	if len(secretKey) == 0 {
		secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))
	}
	return token.SignedString(secretKey)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var loginRequest LoginRequest
	var user models.User
	if err := c.BodyParser(&loginRequest); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid input", err)
	}
	if err := config.DB.Where("username = ? AND user_status = 'active'", loginRequest.Username).First(&user).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusUnauthorized, "Login failed", "User is not active, call admin to activated your account !")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
		return helpers.JSONResponse(c, fiber.StatusUnauthorized, "Login failed", "Invalid username or password")
	}
	token, err := generateJWT(user)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Login failed", "Failed to generate token")
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "Login successful", token)
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	token := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
	if token == "" {
		return helpers.JSONResponse(c, fiber.StatusUnauthorized, "Missing token", "Insert valid token to access this endpoint !")
	}
	if err := blacklistToken(token); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Logout failed", "Failed to blacklist token")
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "Logout successful", "Logout successful")
}

func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	branchID, _ := services.GetBranchID(c)
	userID, _ := services.GetUserID(c)
	userRole, _ := services.GetUserRole(c)
	var profilStruct models.ProfileStruct
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
	return helpers.JSONResponse(c, fiber.StatusOK, "Otoritas : "+userRole, profilStruct)
}

func (h *AuthHandler) GetMenus(c *fiber.Ctx) error {
	userRoles, _ := services.GetUserRole(c)
	data, err := os.ReadFile("menus.json")
	if err != nil {
		log.Printf("Error reading menus.json: %v", err)
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to read menu data", err)
	}
	var menuResponse models.MenuResponse
	if err := json.Unmarshal(data, &menuResponse); err != nil {
		log.Printf("Error unmarshaling menus.json: %v", err)
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to parse menu data", err)
	}
	menus := menuResponse.Data
	userRoleFilter := userRoles
	if userRoleFilter == "" {
		return helpers.JSONResponse(c, fiber.StatusOK, "Get All Menus Success", menus)
	}
	var filteredMenus []models.Menu
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

func (h *AuthHandler) GetBranchByUserId(c *fiber.Ctx) error {
	userID, _ := services.GetUserID(c)
	var userBranchDetails []models.UserBranchDetail
	if err := config.DB.
		Table("user_branches").
		Select("user_branches.user_id, users.name AS user_name, user_branches.branch_id, branches.branch_name, branches.sia_name, branches.sipa_name, branches.phone").
		Joins("LEFT JOIN users ON users.id = user_branches.user_id").
		Joins("LEFT JOIN branches ON branches.id = user_branches.branch_id").
		Where("branches.branch_status = 'active' AND user_branches.user_id = ?", userID).
		Scan(&userBranchDetails).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Get userbranches failed", "Failed to fetch user branches with details")
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "User Branch found", userBranchDetails)
}

func (h *AuthHandler) SetBranch(c *fiber.Ctx) error {
	token := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
	if token == "" {
		return helpers.JSONResponse(c, fiber.StatusUnauthorized, "Missing token", "Insert valid token to access this endpoint!")
	}
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		secretKey := []byte(os.Getenv("JWT_SECRET"))
		if len(secretKey) == 0 {
			secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))
		}
		return secretKey, nil
	})
	if err != nil || !parsedToken.Valid {
		return helpers.JSONResponse(c, fiber.StatusUnauthorized, "Invalid token", "Try to login again!")
	}
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || claims["sub"] == nil {
		return helpers.JSONResponse(c, fiber.StatusUnauthorized, "Invalid token claims", "Try to login again!")
	}
	userID := string(claims["sub"].(string))
	var request struct {
		BranchID string `json:"branch_id" validate:"required"`
	}
	if err := c.BodyParser(&request); err != nil {
		return helpers.JSONResponse(c, fiber.StatusBadRequest, "Invalid input", err)
	}
	var userBranch models.UserBranch
	if err := config.DB.Where("user_id = ? AND branch_id = ?", userID, request.BranchID).First(&userBranch).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusForbidden, "Invalid branch ID", "Branch not associated with this user!")
	}
	var user models.User
	if err := config.DB.Select("name AS name, user_role AS user_role").Where("id = ?", userID).First(&user).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to set branch", "Unable to retrieve user role")
	}
	var branch models.Branch
	if err := config.DB.Select("default_member, quota, subscription_type, real_asset").Where("id = ?", request.BranchID).First(&branch).Error; err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to set branch", "Unable to retrieve branch details")
	}
	newToken, err := generateBranchJWTWithRole(userID, request.BranchID, string(user.UserRole), branch.DefaultMember, branch.Quota, string(branch.SubscriptionType), string(branch.RealAsset), user.Name)
	if err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to set branch", "Failed to generate new token")
	}
	if err := blacklistToken(token); err != nil {
		return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to set branch", "Failed to blacklist old token")
	}
	return helpers.JSONResponse(c, fiber.StatusOK, "Branch set successfully", newToken)
}
