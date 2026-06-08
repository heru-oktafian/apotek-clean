package routes

import (
	handlers "apotek-clean/internal/adapters/driving/http/handlers"
	middlewares "apotek-clean/middlewares"
	fiber "github.com/gofiber/fiber/v2"
)

func RegisterAuditRoutes(app *fiber.App) {
	app.Get("/api/first-stocks", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.GetAllFirstStocks)
	app.Post("/api/first-stocks", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "finance", "superadmin"), handlers.CreateFirstStockTransaction)
	app.Put("/api/first-stocks/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "finance", "superadmin"), handlers.UpdateFirstStock)
	app.Delete("/api/first-stocks/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "finance", "superadmin"), handlers.DeleteFirstStock)
	app.Get("/api/first-stock-with-items/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.GetFirstStockWithItems)
	app.Get("/api/first-stock-items/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.GetAllFirstStockItems)
	app.Post("/api/first-stock-items", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "finance", "superadmin"), handlers.CreateFirstStockItem)
	app.Put("/api/first-stock-items/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "finance", "superadmin"), handlers.UpdateFirstStockItem)
	app.Delete("/api/first-stock-items/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "finance", "superadmin"), handlers.DeleteFirstStockItem)

	app.Get("/api/mobile-opnames", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.GetAllMobileOpnames)
	app.Get("/api/mobile-opnames-active", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.GetAllActiveMobileOpnames)
	app.Get("/api/mobile-opnames-item-details", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.GetMobileOpnameItemDetails)
	app.Get("/api/mobile-opnames-items-glimpse", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.GetMobileOpnameItemsGlimpse)
	app.Get("/api/opnames", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.GetAllOpnames)
	app.Post("/api/opnames", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.CreateOpname)
	app.Get("/api/opnames/:id/items", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.GetOpnameItemsByOpnameID)
	app.Get("/api/opnames/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.GetOpnameWithItems)
	app.Put("/api/opnames/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.UpdateOpnameByID)
	app.Delete("/api/opnames/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.DeleteOpnameByID)
	app.Get("/api/opname-items", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.GetAllOpnameItems)
	app.Post("/api/opname-items-all", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.GetAllOpnameItems)
	app.Post("/api/opname-items", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.CreateOpnameItem)
	app.Put("/api/opname-items", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.UpdateOpnameItemByID)
	app.Delete("/api/opname-items", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.DeleteOpnameItemByID)
	app.Get("/api/cmb-product-opname", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.GetProductsComboboxByName)
}
