package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "github.com/heru-oktafian/fiber-apotek/controllers/masters"
	middlewares "github.com/heru-oktafian/fiber-apotek/middlewares"
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
