package routes

import (
	handlers "apotek-clean/internal/adapters/driving/http/handlers"
	middlewares "apotek-clean/middlewares"
	fiber "github.com/gofiber/fiber/v2"
)

func RegisterBranchRoutes(app *fiber.App) {
	branchHandler := handlers.NewBranchHandler()

	app.Get("/api/branches", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin", "operator", "finance", "cashier"), branchHandler.GetAllBranch)
	app.Get("/api/branches/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin", "operator", "finance", "cashier"), branchHandler.GetBranch)
	app.Post("/api/branches", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin", "operator"), branchHandler.CreateBranch)
	app.Put("/api/branches/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin", "operator"), branchHandler.UpdateBranch)
	app.Delete("/api/branches/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin", "operator"), branchHandler.DeleteBranch)
}
