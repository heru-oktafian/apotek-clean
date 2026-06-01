package routes

import (
	handlers "apotek-clean/internal/adapters/driving/http/handlers"
	middlewares "apotek-clean/middlewares"
	fiber "github.com/gofiber/fiber/v2"
)

func RegisterUserRoutes(app *fiber.App) {
	userHandler := handlers.NewUserHandler()

	app.Get("/api/users", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), userHandler.GetUsers)
	app.Get("/api/users/:user_id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), userHandler.GetUserByID)
	app.Post("/api/users", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), userHandler.CreateUser)
	app.Put("/api/users/:user_id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), userHandler.UpdateUser)
	app.Delete("/api/users/:user_id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), userHandler.DeleteUser)
}
