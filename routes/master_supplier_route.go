package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "apotek-clean/controllers/masters"
	middlewares "apotek-clean/middlewares"
)

func MasterSupplierRoute(app *fiber.App) {
	// Supplier routes
	app.Get("/api/suppliers", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetAllSupplier)
	app.Get("/api/suppliers/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetSupplierByID)
	app.Post("/api/suppliers", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.CreateSupplier)
	app.Put("/api/suppliers/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.UpdateSupplier)
	app.Delete("/api/suppliers/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.DeleteSupplier)

	// Supplier Combobox route
	app.Get("/api/suppliers-combo", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.CmbSupplier)
}
