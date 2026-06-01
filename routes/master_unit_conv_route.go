package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "github.com/heru-oktafian/fiber-apotek/controllers/masters"
	middlewares "github.com/heru-oktafian/fiber-apotek/middlewares"
)

func MasterUnitConvRoutes(app *fiber.App) {
	// Endpoint Unit Conversion
	app.Get("/api/unit-conversions", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin", "operator", "finance", "cashier"), controllers.GetAllUnitConversion)
	app.Get("/api/unit-conversions/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin", "operator", "finance", "cashier"), controllers.GetUnitConversionByID)
	app.Post("/api/unit-conversions", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.CreateUnitConversion)
	app.Put("/api/unit-conversions/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.UpdateUnitConversion)
	app.Delete("/api/unit-conversions/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.DeleteUnitConversion)

	app.Get("/api/conversion-products-combo", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin", "operator", "finance", "cashier"), controllers.CmbProdConv)
	app.Get("/api/units-combo", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin", "operator", "finance", "cashier"), controllers.CmbUnit)
}
