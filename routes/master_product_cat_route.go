package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "apotek-clean/controllers/masters"
	middlewares "apotek-clean/middlewares"
)

func MasterProductCatRoute(app *fiber.App) {
	// Endpoint Product Categories
	app.Get("/api/product-categories", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetAllProductCategory)
	app.Post("/api/product-categories", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.CreateProductCategory)
	app.Get("/api/product-categories/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetProductCategory)
	app.Put("/api/product-categories/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.UpdateProductCategory)
	app.Delete("/api/product-categories/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.DeleteProductCategory)

	// Endpoint Product Category Combobox
	app.Get("/api/product-categories-combo", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.CmbProductCategory)
}
