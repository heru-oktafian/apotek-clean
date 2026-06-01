package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "apotek-clean/controllers/systems"
	middlewares "apotek-clean/middlewares"
)

func SysDefectaRoute(app *fiber.App) {
	// Defecta routes
	app.Get("/api/sys-defectas", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetAllDefectas)
	app.Get("/api/sys-defectas/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetDefetaWithItems)
	app.Post("/api/sys-defectas", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.CreateDefecta)
	app.Put("/api/sys-defectas/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.UpdateDefecta)
	app.Delete("/api/sys-defectas/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.DeleteDefecta)

	// Defecta items routes
	app.Get("/api/sys-defecta-items", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetAllDefectaItems)
	app.Post("/api/sys-defecta-items", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.CreateDefectaItem)
	app.Put("/api/sys-defecta-items/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.UpdateDefectaItem)
	app.Delete("/api/sys-defecta-items/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.DeleteDefectaItem)
}
