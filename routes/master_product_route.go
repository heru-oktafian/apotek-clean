package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "github.com/heru-oktafian/fiber-apotek/controllers/masters"
	middlewares "github.com/heru-oktafian/fiber-apotek/middlewares"
)

func MasterProductRoute(app *fiber.App) {
	// Endpoint Products
	app.Get("/api/products", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetAllProduct)
	app.Post("/api/products", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.CreateProduct)
	app.Get("/api/products/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetProduct)
	app.Put("/api/products/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.UpdateProduct)
	app.Delete("/api/products/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.DeleteProduct)

	// Endpoint Product Combobox untuk penjualan
	app.Get("/api/sales-products-combo", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.CmbProdSale)

	// Endpoint Product Combobox untuk pembelian
	app.Get("/api/purchase-products-combo", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.CmbProdPurchase)
}
