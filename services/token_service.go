package services

import (
	fmt "fmt"
	os "os"
	strings "strings"

	fiber "github.com/gofiber/fiber/v2"
	jwt "github.com/golang-jwt/jwt/v5"
)

// GetUserID mendapatkan user_id dari token
func GetUserID(c *fiber.Ctx) (string, error) {
	// Dapatkan user_id dari klaim
	userID, _ := GetClaimsToken(c, "sub")
	// println("+" + userID)
	return userID, nil
}

// GetBranchID mendapatkan branch_id dari token
func GetBranchID(c *fiber.Ctx) (string, error) {
	// Dapatkan branch_id dari klaim
	branchID, _ := GetClaimsToken(c, "branch_id")

	return branchID, nil
}

// GetUserRole mendapatkan user_role dari token
func GetUserRole(c *fiber.Ctx) (string, error) {
	// Dapatkan user_role dari klaim
	userRole, _ := GetClaimsToken(c, "user_role")

	return userRole, nil
}

// GetDefaultMember mendapatkan default_member dari token
func GetDefaultMember(c *fiber.Ctx) (string, error) {
	defaultMember, _ := GetClaimsToken(c, "default_member")

	return defaultMember, nil
}

// GetClaimsToken mendapatkan nilai klaim dari token
func GetClaimsToken(c *fiber.Ctx, key string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("context is nil")
	}
	// Ambil nilai token dari header Authorization
	authHeader := c.Get("Authorization")
	// fmt.Println("Authorization Header:", authHeader)

	token := authHeader
	// Hapus prefix "Bearer " jika ada
	token = strings.TrimPrefix(token, "Bearer ")
	// fmt.Println("Stripped Token:", token)

	// Periksa apakah token kosong
	if token == "" {
		return "", fmt.Errorf("missing token")
	}

	// Verifikasi token JWT
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		secretKey := []byte(os.Getenv("JWT_SECRET"))
		return secretKey, nil
	})

	if err != nil {
		// fmt.Println("Token parse error:", err)
		return "", fmt.Errorf("invalid token")
	}

	if !parsedToken.Valid {
		// fmt.Println("Token is not valid")
		return "", fmt.Errorf("invalid token")
	}

	// Dapatkan klaim dari token
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		// fmt.Println("Failed to cast claims to MapClaims")
		return "", fmt.Errorf("invalid token claims")
	}
	// fmt.Printf("All Claims: %+v\n", claims)

	// Dapatkan nilai dari klaim
	claimValRaw, ok := claims[key]
	if !ok || claimValRaw == nil {
		// fmt.Println("Claim for key", key, "not found")
		return "", fmt.Errorf("claim '%s' not found in token", key)
	}

	claimedValue := fmt.Sprintf("%v", claimValRaw)
	// fmt.Println("Claimed value for key", key, ":", claimedValue)

	return claimedValue, nil
}
