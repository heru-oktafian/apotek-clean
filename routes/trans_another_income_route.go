package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "github.com/heru-oktafian/fiber-apotek/controllers/transactions"
	middlewares "github.com/heru-oktafian/fiber-apotek/middlewares"
)

func TransAnotherIncomeRoute(app *fiber.App) {
	app.Get("/api/another-incomes", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetAllAnotherIncomes)
	app.Post("/api/another-incomes/", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.CreateAnotherIncome)
	app.Put("/api/another-incomes/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.UpdateAnotherIncome)
	app.Delete("/api/another-incomes/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.DeleteAnotherIncome)
}
