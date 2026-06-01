package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	legacyroutes "apotek-clean/routes"
)

func RegisterExportRoutes(app *fiber.App) {
	legacyroutes.ExportExcelRoutes(app)
	legacyroutes.ExportPDFRoutes(app)
}
