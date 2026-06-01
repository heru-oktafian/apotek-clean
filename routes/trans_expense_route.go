package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "apotek-clean/controllers/transactions"
	middlewares "apotek-clean/middlewares"
)

func TransExpenseRoutes(app *fiber.App) {
	app.Post("/api/expenses", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.CreateExpense)
	app.Get("/api/expenses", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.GetAllExpenses)
	app.Put("/api/expenses/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.UpdateExpense)
	app.Delete("/api/expenses/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.DeleteExpense)
}
