package helpers

import (
	fmt "fmt"
	log "log"
	os "os"
	strings "strings"

	fiber "github.com/gofiber/fiber/v2"
	jwt "github.com/golang-jwt/jwt/v5"
	configs "github.com/heru-oktafian/fiber-apotek/configs"
)

// TokenValidation memvalidasi token
//
// Behavior berubah berdasarkan variabel lingkungan REDIS_AUTH:
//   - TRUE: cek blacklist di Redis
//   - FALSE: hanya verifikasi expiry tanpa blacklist

func TokenValidation(c *fiber.Ctx, key string) error {
	// Ambil nilai token dari header Authorization
	token := c.Get("Authorization")
	// Hapus awalan "Bearer " jika ada
	token = strings.TrimPrefix(token, "Bearer ")

	// Periksa apakah token kosong
	if token == "" {
		return JSONResponse(c, fiber.StatusUnauthorized, "Missing token", "Insert valid token to access this endpoint!")
	}

	// Tentukan apakah pengecekan blacklist Redis aktif lewat .env
	// default ke FALSE bila variabel tidak diatur atau bernilai selain "TRUE"
	redisAuth := strings.ToUpper(os.Getenv("REDIS_AUTH")) == "TRUE"
	if redisAuth {
		// Periksa token di daftar hitam Redis hanya ketika redisAuth = TRUE
		redisKey := fmt.Sprintf("blacklist:%s", token)
		rdb := configs.RDB
		isBlacklisted, err := rdb.Exists(configs.Ctx, redisKey).Result()

		if err != nil {
			log.Printf("Error checking token in Redis: %v", err)
			return JSONResponse(c, fiber.StatusInternalServerError, "Token verification failed", "Server error!")
		}

		if isBlacklisted > 0 {
			return JSONResponse(c, fiber.StatusUnauthorized, "Using token failed", "Token was revoked, please login again!")
		}
	}

	// Verifikasi token menggunakan kunci rahasia
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		secretKey := []byte(os.Getenv("JWT_SECRET"))
		return secretKey, nil
	})

	if err != nil || !parsedToken.Valid {
		return JSONResponse(c, fiber.StatusUnauthorized, "Invalid token", "Try to login again!")
	}

	// Periksa klaim token (opsional, misalnya validasi user_id, role, dll.)
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || claims[key] == nil {
		return JSONResponse(c, fiber.StatusUnauthorized, "Invalid token claims", "Try to login again!")
	}

	// Lanjut ke middleware berikutnya
	return c.Next()
}
