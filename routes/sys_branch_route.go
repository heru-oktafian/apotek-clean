package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "apotek-clean/controllers/systems"
	middlewares "apotek-clean/middlewares"
)

func SysBranchRoutes(app *fiber.App) {
	// Endpoint Cabang
	app.Get("/api/branches", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin", "operator", "finance", "cashier"), controllers.GetAllBranch)
	app.Get("/api/branches/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin", "operator", "finance", "cashier"), controllers.GetBranch)
	app.Post("/api/branches", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin", "operator"), controllers.CreateBranch)
	app.Put("/api/branches/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin", "operator"), controllers.UpdateBranch)
	app.Delete("/api/branches/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin", "operator"), controllers.DeleteBranch)
}
