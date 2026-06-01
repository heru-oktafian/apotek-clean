package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "github.com/heru-oktafian/fiber-apotek/controllers/systems"
	middlewares "github.com/heru-oktafian/fiber-apotek/middlewares"
)

func SysUserRoute(app fiber.Router) {
	app.Get("/api/users", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetUsers)
	app.Get("/api/users/:user_id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetUserByID)
	app.Post("/api/users", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.CreateUser)
	app.Put("/api/users/:user_id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.UpdateUser)
	app.Delete("/api/users/:user_id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.DeleteUser)
}
