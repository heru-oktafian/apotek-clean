package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "github.com/heru-oktafian/fiber-apotek/controllers/transactions"
	middlewares "github.com/heru-oktafian/fiber-apotek/middlewares"
)

func TransExpenseRoutes(app *fiber.App) {
	app.Post("/api/expenses", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.CreateExpense)
	app.Get("/api/expenses", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.GetAllExpenses)
	app.Put("/api/expenses/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.UpdateExpense)
	app.Delete("/api/expenses/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.DeleteExpense)
}
