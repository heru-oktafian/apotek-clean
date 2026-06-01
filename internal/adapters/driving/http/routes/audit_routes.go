package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	legacyroutes "apotek-clean/routes"
)

func RegisterAuditRoutes(app *fiber.App) {
	legacyroutes.AudFirstStockRoutes(app)
	legacyroutes.AudOpnameRoute(app)
}
