package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "apotek-clean/controllers/transactions"
	middlewares "apotek-clean/middlewares"
)

func TransAnotherIncomeRoute(app *fiber.App) {
	app.Get("/api/another-incomes", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetAllAnotherIncomes)
	app.Post("/api/another-incomes/", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.CreateAnotherIncome)
	app.Put("/api/another-incomes/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.UpdateAnotherIncome)
	app.Delete("/api/another-incomes/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.DeleteAnotherIncome)
}
