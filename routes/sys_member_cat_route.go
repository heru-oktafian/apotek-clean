package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "github.com/heru-oktafian/fiber-apotek/controllers/systems"
	middlewares "github.com/heru-oktafian/fiber-apotek/middlewares"
)

func SysMemberCatRoute(app *fiber.App) {
	app.Get("/api/member-categories", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetAllMemberCategory)
	app.Get("/api/member-categories/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetMemberCategory)
	app.Post("/api/member-categories", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.CreateMemberCategory)
	app.Put("/api/member-categories/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.UpdateMemberCategory)
	app.Delete("/api/member-categories/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.DeleteMemberCategory)

	app.Get("/api/member-categories-combo", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.CmbMemberCategory)
}
