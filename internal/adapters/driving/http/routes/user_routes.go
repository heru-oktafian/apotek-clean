package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	legacyroutes "apotek-clean/routes"
)

func RegisterUserRoutes(app *fiber.App) {
	legacyroutes.SysUserRoute(app)
}
