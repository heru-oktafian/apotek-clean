package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "github.com/heru-oktafian/fiber-apotek/controllers/systems"
	middlewares "github.com/heru-oktafian/fiber-apotek/middlewares"
)

func SysMemberRoute(app *fiber.App) {
	app.Get("/api/members", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetAllMember)
	app.Get("/api/members/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetMember)
	app.Post("/api/members", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.CreateMember)
	app.Put("/api/members/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.UpdateMember)
	app.Delete("/api/members/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.DeleteMember)

	app.Get("/api/members-combo", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.CmbMember)
}
