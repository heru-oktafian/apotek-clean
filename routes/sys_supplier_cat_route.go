package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "apotek-clean/controllers/systems"
	middlewares "apotek-clean/middlewares"
)

func SysSupplierCatRoute(app *fiber.App) {
	// Supplier Category routes
	app.Get("/api/supplier-categories", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetAllSupplierCategory)

	app.Delete("/api/supplier-categories/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.DeleteSupplierCategory)

	// Supplier Category Combobox route
	app.Get("/api/supplier-categories-combo", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.CmbSupplierCategory)
}
