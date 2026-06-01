package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	legacyroutes "apotek-clean/routes"
)

func RegisterMasterDataRoutes(app *fiber.App) {
	legacyroutes.MasterProductCatRoute(app)
	legacyroutes.MasterProductRoute(app)
	legacyroutes.MasterSupplierRoute(app)
	legacyroutes.MasterUnitRoutes(app)
	legacyroutes.MasterUnitConvRoutes(app)
	legacyroutes.SysMemberCatRoute(app)
	legacyroutes.SysMemberRoute(app)
	legacyroutes.SysSupplierCatRoute(app)
}
