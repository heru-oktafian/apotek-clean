package routes

import (
	os "os"

	fiber "github.com/gofiber/fiber/v2"
	controllers "github.com/heru-oktafian/fiber-apotek/controllers/systems"
	helpers "github.com/heru-oktafian/fiber-apotek/helpers"
	middlewares "github.com/heru-oktafian/fiber-apotek/middlewares"
	services "github.com/heru-oktafian/fiber-apotek/services"
)

func AuthRoutes(app *fiber.App) {

	app.Get("/coba", func(c *fiber.Ctx) error {
		// defaultMember := helpers.GetMemberDefault(config.DB, "BRC250118132203")
		// // return defaultMember
		// return c.SendString(defaultMember)
		return c.SendString("Halo dari Fiber di port " + os.Getenv("SERVER_PORT"))
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

	// Endpoint Menu
	app.Get("/api/menus", middlewares.JWTMiddleware, controllers.GetMenus)

	// Endpoint Otentikasi
	app.Post("/api/login", controllers.Login)
	app.Post("/api/logout", controllers.Logout)
	app.Get("/api/profile", middlewares.JWTMiddleware, controllers.GetProfile)
	app.Get("/api/list_branches", middlewares.JWTMiddleware, controllers.GetBranchByUserId)

	// Endpoint SetBranch
	app.Post("/api/set_branch", controllers.SetBranch)

	// Endpoint untuk menghasilkan file .env
	app.Post("/api/update-env", middlewares.JWTMiddleware, middlewares.RoleMiddleware("superadmin", "administrator"), func(c *fiber.Ctx) error {
		type request struct {
			Content string `json:"content"`
		}

		var body request
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Gagal membaca request body",
			})
		}

		if body.Content == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Content tidak boleh kosong",
			})
		}

		if err := services.WriteRawEnvFile(".env", body.Content); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "✅ File .env berhasil diperbarui",
		})
	})
}
