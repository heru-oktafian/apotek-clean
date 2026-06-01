package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "github.com/heru-oktafian/fiber-apotek/controllers/systems"
	middlewares "github.com/heru-oktafian/fiber-apotek/middlewares"
)

func SysUserBranchRoutes(app *fiber.App) {
	// UserBranch routes
	app.Get("/api/user-branches", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetAllUserBranch)
	app.Get("/api/user-branches/:user_id/:branch_id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetUserBranch)
	app.Post("/api/user-branches", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.CreateUserBranch)
	app.Put("/api/user-branches/:user_id/:branch_id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.UpdateUserBranch)
	app.Delete("/api/user-branches/:user_id/:branch_id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.DeleteUserBranch)
	app.Get("/api/detail-users/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetUserDetails)
}
