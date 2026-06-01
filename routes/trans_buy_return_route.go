package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "apotek-clean/controllers/transactions"
	middlewares "apotek-clean/middlewares"
)

func TransBuyReturnRoutes(app *fiber.App) {
	// Rute untuk transaksi retur pembelian
	app.Get("/api/buy-returns", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetAllBuyReturns)
	app.Post("/api/buy-returns", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.CreateBuyReturnTransaction)
	app.Get("/api/buy-returns/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetBuyReturnWithItems)

	// Rute untuk combo box
	app.Get("/api/cmb-prod-buy-returns", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetBuyItemsForReturn)
	app.Get("/api/cmb-purchases", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.CmbPurchase)
}
