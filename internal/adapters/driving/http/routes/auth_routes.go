package routes

import (
	os "os"

	handlers "apotek-clean/internal/adapters/driving/http/handlers"
	helpers "apotek-clean/helpers"
	middlewares "apotek-clean/middlewares"
	services "apotek-clean/services"
	fiber "github.com/gofiber/fiber/v2"
)

func RegisterAuthRoutes(app *fiber.App) {
	authHandler := handlers.NewAuthHandler()

	app.Get("/coba", func(c *fiber.Ctx) error {
		port := os.Getenv("PORT")
		if port == "" {
			port = os.Getenv("APP_PORT")
		}
		if port == "" {
			port = os.Getenv("SERVER_PORT")
		}
		return c.SendString("Halo dari Fiber di port " + port)
	})

	app.Post("/", func(c *fiber.Ctx) error {
		return helpers.JSONResponse(c, fiber.StatusOK, "Pesan anda telah kami terima dan segera kami tindak lanjuti.", nil)
	})

	app.Get("/files/dump", middlewares.JWTMiddleware, middlewares.RoleMiddleware("superadmin", "administrator"), func(c *fiber.Ctx) error {
		files, err := services.ListDumpFiles()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(files)
	})

	app.Get("/files/rest", middlewares.JWTMiddleware, middlewares.RoleMiddleware("superadmin", "administrator"), func(c *fiber.Ctx) error {
		files, err := services.ListRestFiles()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(files)
	})

	app.Get("/api/menus", middlewares.JWTMiddleware, authHandler.GetMenus)
	app.Post("/api/login", authHandler.Login)
	app.Post("/api/logout", authHandler.Logout)
	app.Get("/api/profile", middlewares.JWTMiddleware, authHandler.GetProfile)
	app.Get("/api/list_branches", middlewares.JWTMiddleware, authHandler.GetBranchByUserId)
	app.Post("/api/set_branch", authHandler.SetBranch)

	app.Post("/api/update-env", middlewares.JWTMiddleware, middlewares.RoleMiddleware("superadmin", "administrator"), func(c *fiber.Ctx) error {
		type request struct {
			Content string `json:"content"`
		}

		var body request
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Gagal membaca request body"})
		}

		if body.Content == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Content tidak boleh kosong"})
		}

		if err := services.WriteRawEnvFile(".env", body.Content); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{"message": "✅ File .env berhasil diperbarui"})
	})
}
