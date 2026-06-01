package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	legacyroutes "apotek-clean/routes"
)

func RegisterBranchRoutes(app *fiber.App) {
	legacyroutes.SysBranchRoutes(app)
}
