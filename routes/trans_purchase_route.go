package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "github.com/heru-oktafian/fiber-apotek/controllers/transactions"
	middlewares "github.com/heru-oktafian/fiber-apotek/middlewares"
)

func TransPurchaseRoutes(app *fiber.App) {
	// Purchase Routes
	app.Get("/api/purchases", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.GetAllPurchases)
	app.Get("/api/purchases/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.GetPurchaseWithItems)
	app.Post("/api/purchases", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.CreatePurchaseTransaction)
	app.Put("/api/purchases/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.UpdatePurchase)
	app.Delete("/api/purchases/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.DeletePurchase)

	app.Get("/api/purchase-items/all/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.GetAllPurchaseItems)
	app.Post("/api/purchase-items", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.CreatePurchaseItem)
	app.Put("/api/purchase-items/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.UpdatePurchaseItem)
	app.Delete("/api/purchase-items/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.DeletePurchaseItem)
}
