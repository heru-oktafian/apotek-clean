package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "github.com/heru-oktafian/fiber-apotek/controllers/audits"
	middlewares "github.com/heru-oktafian/fiber-apotek/middlewares"
)

func AudFirstStockRoutes(app *fiber.App) {
	// Endpoint First Stock
	app.Get("/api/first-stocks", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetAllFirstStocks)
	app.Post("/api/first-stocks", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "finance", "superadmin"), controllers.CreateFirstStockTransaction)
	app.Put("/api/first-stocks/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "finance", "superadmin"), controllers.UpdateFirstStock)
	app.Delete("/api/first-stocks/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "finance", "superadmin"), controllers.DeleteFirstStock)

	// Endpoint First Stock With Items Detail
	app.Get("/api/first-stock-with-items/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetFirstStockWithItems)

	// Endpoint First Stock Items
	app.Get("/api/first-stock-items/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetAllFirstStockItems)
	app.Post("/api/first-stock-items", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "finance", "superadmin"), controllers.CreateFirstStockItem)
	app.Put("/api/first-stock-items/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "finance", "superadmin"), controllers.UpdateFirstStockItem)
	app.Delete("/api/first-stock-items/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "finance", "superadmin"), controllers.DeleteFirstStockItem)

}
