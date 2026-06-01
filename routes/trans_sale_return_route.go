package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "github.com/heru-oktafian/fiber-apotek/controllers/transactions"
	middlewares "github.com/heru-oktafian/fiber-apotek/middlewares"
)

func TransSaleReturnRoutes(app *fiber.App) {
	// Sale Return Routes
	app.Get("/api/sale-returns", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.GetAllSaleReturns)
	app.Get("/api/sale-returns/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.GetSaleReturnWithItems)
	app.Post("/api/sale-returns", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.CreateSaleReturnTransaction)

	// Combobox Sale Return Routes
	app.Get("/api/cmb-prod-sale-returns", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.GetSaleItemsForReturn)

	// Combobox Sale Routes
	app.Get("/api/cmb-sales", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.CmbSale)
}
