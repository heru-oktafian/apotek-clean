package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "github.com/heru-oktafian/fiber-apotek/controllers/audits"
	middlewares "github.com/heru-oktafian/fiber-apotek/middlewares"
)

func AudOpnameRoute(app *fiber.App) {
	// Endpoint Mobile Opnames
	app.Get("/api/mobile-opnames", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetAllMobileOpnames)
	app.Get("/api/mobile-opnames-active", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetAllActiveMobileOpnames)
	app.Get("/api/mobile-opnames-item-details", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetMobileOpnameItemDetails)
	app.Get("/api/mobile-opnames-items-glimpse", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetMobileOpnameItemsGlimpse)

	// Endpoint Opnames
	app.Get("/api/opnames", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetAllOpnames)
	app.Post("/api/opnames", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.CreateOpname)
	app.Get("/api/opnames/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetOpnameWithItems)
	app.Put("/api/opnames/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.UpdateOpnameByID)
	app.Delete("/api/opnames/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.DeleteOpnameByID)

	// Endpoint Opname Items (ID now supplied in request body)
	app.Get("/api/opname-items", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetAllOpnameItems)
	app.Post("/api/opname-items-all", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetAllOpnameItems)
	app.Post("/api/opname-items", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.CreateOpnameItem)
	app.Put("/api/opname-items", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.UpdateOpnameItemByID)
	app.Delete("/api/opname-items", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.DeleteOpnameItemByID)

	// Endpoint Products Combobox
	app.Get("/api/cmb-product-opname", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetProductsComboboxByName)
}
