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
	godotenv "github.com/joho/godotenv"

	configs "apotek-clean/configs"
	helpers "apotek-clean/helpers"
	seeders "apotek-clean/seeders"
	crons "apotek-clean/services/crons"
	internalroutes "apotek-clean/internal/adapters/driving/http/routes"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	configs.InitTimezone()
	log.Println("🕒 Sekarang WIB:", time.Now().In(configs.Location))

	serverPort := os.Getenv("APP_PORT")
	if serverPort == "" {
		serverPort = os.Getenv("PORT")
	}
	if serverPort == "" {
		serverPort = os.Getenv("SERVER_PORT")
	}
	if serverPort == "" {
		serverPort = "9001"
	}

	if err := configs.SetupDB(); err != nil {
		log.Fatal(err)
	}

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

	go func() {
		crons.SchedulerJobs(configs.DB)
	}()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))
	app.Use(limiter.New(limiter.Config{
		Max:        80,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return helpers.JSONResponse(c, fiber.StatusTooManyRequests, "Terlalu banyak permintaan (Rate limit tercapai). Silakan coba lagi nanti.", nil)
		},
	}))

	internalroutes.RegisterCoreRoutes(app)

	routeCount := 0
	for _, routes := range app.Stack() {
		routeCount += len(routes)
	}

	port, err := strconv.Atoi(serverPort)
	if err != nil {
		log.Fatal("Invalid APP_PORT/SERVER_PORT: must be a number")
	}

	helpers.PrintFiberLikeBanner(
		os.Getenv("APPNAME"),
		"0.0.0.0",
		port,
		routeCount,
	)

	log.Fatal(app.Listen(":" + serverPort))
}
