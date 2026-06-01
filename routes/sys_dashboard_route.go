package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "github.com/heru-oktafian/fiber-apotek/controllers/systems"
	middlewares "github.com/heru-oktafian/fiber-apotek/middlewares"
)

func SysDashboardRoute(app *fiber.App) {
	app.Get("/api/dashboard/monthly-profit-report", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.MonthlyProfitReport)
	app.Get("/api/dashboard/daily-profit-report", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.DailyProfitReport)
	app.Get("/api/dashboard/weekly-profit-report", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.WeeklyProfitReport)
	app.Get("/api/dashboard/profit-today-by-user", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetDailyProfitReportByUser)
	app.Get("/api/dashboard/top-selling-report", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetTopSellingProducts)
	app.Get("/api/dashboard/least-selling-report", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetLeastSellingProducts)
	app.Get("/api/dashboard/neared-report", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetExpiringProducts)
}
