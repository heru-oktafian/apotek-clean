package main

import (
	log "log"
	os "os"
	strconv "strconv"
	time "time"

	fiber "github.com/gofiber/fiber/v2"
	cors "github.com/gofiber/fiber/v2/middleware/cors"
	limiter "github.com/gofiber/fiber/v2/middleware/limiter"
	logger "github.com/gofiber/fiber/v2/middleware/logger"
	configs "github.com/heru-oktafian/fiber-apotek/configs"
	helpers "github.com/heru-oktafian/fiber-apotek/helpers"
	routes "github.com/heru-oktafian/fiber-apotek/routes"
	seeders "github.com/heru-oktafian/fiber-apotek/seeders"
	crons "github.com/heru-oktafian/fiber-apotek/services/crons"
	godotenv "github.com/joho/godotenv"
)

func main() {
	// Muat file .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Inisialisasi zona waktu global
	configs.InitTimezone()
	log.Println("🕒 Sekarang WIB:", time.Now().In(configs.Location))

	// Dapatkan port dari environment
	serverPort := os.Getenv("SERVER_PORT")

	// Inisialisasi database
	if err := configs.SetupDB(); err != nil {
		log.Fatal(err)
	}

	// Jalankan seeder jika ada argumen "seed"
	if len(os.Args) > 1 && os.Args[1] == "seed" {
		seeders.UserSeed()
		seeders.BranchSeed()
		seeders.UserBranchSeed()
		seeders.UnitSeed()
		seeders.UnitConversionSeed()
		seeders.ProductCategorySeed()
		seeders.ProductSeed()
		seeders.MemberCategorySeed()
		seeders.SupplierCategorySeed()
		seeders.SupplierSeed()
		os.Exit(0)
	}

	// Jalankan routine
	go func() {
		// Jalankan scheduler jobs
		crons.SchedulerJobs(configs.DB)
	}()

	// Inisialisasi Fiber
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true, // ⛔ wajib
	})

	// Menambahkan middleware logger fiber
	app.Use(logger.New())

	// Aktifkan CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // Sesuaikan jika kamu ingin membatasi domain tertentu
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Aktifkan Rate Limiter
	app.Use(limiter.New(limiter.Config{
		Max:        80,              // Maksimal 80 request
		Expiration: 1 * time.Minute, // Tiap 1 menit
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() // Membatasi berdasarkan alamat IP
		},
		LimitReached: func(c *fiber.Ctx) error {
			return helpers.JSONResponse(c, fiber.StatusTooManyRequests, "Terlalu banyak permintaan (Rate limit tercapai). Silakan coba lagi nanti.", nil)
		},
	}))

	// Setup rute
	routes.AuthRoutes(app)
	routes.ExportExcelRoutes(app) // ✅ Harus sebelum MasterProductRoute agar tidak ter-match oleh :id pattern
	routes.ExportPDFRoutes(app)
	routes.SysBranchRoutes(app)
	routes.AudFirstStockRoutes(app)
	routes.MasterProductCatRoute(app)
	routes.MasterProductRoute(app)
	routes.SysSupplierCatRoute(app)
	routes.MasterSupplierRoute(app)
	routes.MasterUnitRoutes(app)
	routes.MasterUnitConvRoutes(app)
	routes.SysDashboardRoute(app)
	routes.SysDailyAssetRoute(app)
	routes.AudOpnameRoute(app)
	routes.SysDefectaRoute(app)
	routes.SysMemberCatRoute(app)
	routes.SysMemberRoute(app)
	routes.SysReportRoute(app)
	routes.SysUserRoute(app)
	routes.SysUserBranchRoutes(app)
	routes.TransAnotherIncomeRoute(app)
	routes.TransBuyReturnRoutes(app)
	routes.TransDuplicateReceiptRoutes(app)
	routes.TransExpenseRoutes(app)
	routes.TransPurchaseRoutes(app)
	routes.TransSaleRoutes(app)
	routes.TransSaleReturnRoutes(app)

	// Hitung total rute
	routeCount := 0
	for _, routes := range app.Stack() {
		routeCount += len(routes)
	}

	// Konversi port ke integer
	port, err := strconv.Atoi(serverPort)
	if err != nil {
		log.Fatal("Invalid SERVER_PORT: must be a number")
	}

	// Tampilkan banner Fiber
	helpers.PrintFiberLikeBanner(
		os.Getenv("APPNAME"),
		"0.0.0.0",
		port,
		routeCount, // jumlah handlers
	)

	// Jalankan server Fiber
	log.Fatal(app.Listen(":" + serverPort))
}
