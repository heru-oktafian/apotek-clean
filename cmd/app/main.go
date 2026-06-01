package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

	// Sesuaikan path import ini berdasarkan struktur proyek Anda
	// Jika configs ada di internal/configs, gunakan path ini:
	configs "apotek-clean/internal/configs" 
	
	// Adapters
	driving_http "apotek-clean/internal/adapters/driving/http"
	// ... import adapter driven (repository) yang akan di-inject
	
	// Core
	"apotek-clean/internal/core/usecases"
	// ... import entities, ports jika diperlukan di DI

	// Frameworks
	"apotek-clean/internal/frameworks/database"
	"apotek-clean/internal/frameworks/web"
	"apotek-clean/internal/frameworks/auth"
	// ... import frameworks lain jika ada
)

func main() {
	// 1. Load .env file
	err := godotenv.Load() // Memuat file .env di root direktori
	if err != nil {
		log.Println("Warning: Could not load .env file. Using default configurations or environment variables.")
	}

	// 2. Load Configurations
	dbConfig := configs.LoadDatabaseConfig()
	serverConfig := configs.LoadServerConfig() // Asumsi ada LoadServerConfig untuk APP_PORT dll.

	// 3. Initialize Database Connection
	db, err := database.NewPostgresConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: ¥v", err)
	}
	defer db.Close()
	// TODO: Mungkin perlu Migrate DB di sini jika belum otomatis

	// 4. Initialize Repositories (Driven Adapters)
	//    Ini adalah bagian dari Dependency Injection
	//    Kita akan membuat instance dari repository yang menggunakan koneksi DB
	//    Contoh:
	//    productRepo := postgres.NewProductRepository(db)
	//    userRepo := postgres.NewUserRepository(db)
	//    ... dll

	// 5. Initialize Token Service (jika terpisah dari DB, contoh: misal pakai JWT secret dari config)
	//    tokenService := auth.NewJWTService("your_secret_key_from_env") // Adaptasi ini
	
	// 6. Initialize Use Cases (Core)
	//    Use cases akan menerima instance repository sebagai argumen
	//    Contoh:
	//    userUC := usecases.NewUserService(userRepo)
	//    productUC := usecases.NewProductService(productRepo)
	//    authUC := usecases.NewAuthService(userRepo, tokenService)
	//    ... dll

	// 7. Initialize HTTP Handlers (Driving Adapters)
	//    Handlers akan menerima instance Use Cases sebagai argumen
	//    Contoh:
	//    productHandler := driving_http.NewProductHandler(productUC)
	//    userHandler := driving_http.NewUserHandler(userUC)
	//    authHandler := driving_http.NewAuthHandler(authUC)
	//    ... dll

	// 8. Initialize Fiber Web Server
	//    Server config diambil dari serverConfig (APP_PORT)
	app := web.NewFiberServer(serverConfig.Port) // Asumsi ada fungsi NewFiberServer di internal/frameworks/web

	// Muat Routers dan Middleware
	// Di sini akan dipanggil fungsi setup router dari internal/adapters/driving/http/routes/
	// Contoh:
	// app.Use(logger.New()) // Middleware umum
	// auth.SetupRoutes(app, authHandler) // Setup routes auth
	// product_routes.SetupRoutes(app, productHandler) // Setup routes product
	// ... dll

	// 9. Start Server
	log.Printf("Server starting on port %s...", serverConfig.Port)
	if err := app.Listen(serverConfig.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// TODO:
// - Implement Dependency Injection properly (e.g., using a DI container or manual wiring in main)
// - Fill in the actual repository implementations in internal/adapters/driven/postgres/
// - Adjust import paths if internal/configs or other modules are located elsewhere.
// - Implement proper error handling and logging.
// - Add necessary migrations or seeding if needed.
