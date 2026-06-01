package routes

import (
	handlers "apotek-clean/internal/adapters/driving/http/handlers"
	middlewares "apotek-clean/middlewares"
	fiber "github.com/gofiber/fiber/v2"
)

func RegisterSystemRoutes(app *fiber.App) {
	userBranchHandler := handlers.NewUserBranchHandler()
	dashboardHandler := handlers.NewDashboardHandler()
	dailyAssetHandler := handlers.NewDailyAssetHandler()
	defectaHandler := handlers.NewDefectaHandler()
	reportHandler := handlers.NewReportHandler()

	app.Get("/api/user-branches", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), userBranchHandler.GetAllUserBranch)
	app.Get("/api/user-branches/:user_id/:branch_id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), userBranchHandler.GetUserBranch)
	app.Post("/api/user-branches", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), userBranchHandler.CreateUserBranch)
	app.Put("/api/user-branches/:user_id/:branch_id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), userBranchHandler.UpdateUserBranch)
	app.Delete("/api/user-branches/:user_id/:branch_id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), userBranchHandler.DeleteUserBranch)
	app.Get("/api/detail-users/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), userBranchHandler.GetUserDetails)

	app.Get("/api/dashboard/monthly-profit-report", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), dashboardHandler.MonthlyProfitReport)
	app.Get("/api/dashboard/daily-profit-report", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), dashboardHandler.DailyProfitReport)
	app.Get("/api/dashboard/weekly-profit-report", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), dashboardHandler.WeeklyProfitReport)
	app.Get("/api/dashboard/profit-today-by-user", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), dashboardHandler.GetDailyProfitReportByUser)
	app.Get("/api/dashboard/top-selling-report", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), dashboardHandler.GetTopSellingProducts)
	app.Get("/api/dashboard/least-selling-report", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), dashboardHandler.GetLeastSellingProducts)
	app.Get("/api/dashboard/neared-report", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), dashboardHandler.GetExpiringProducts)

	app.Get("/api/daily_asset", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), dailyAssetHandler.GetAllAssets)

	app.Get("/api/sys-defectas", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), defectaHandler.GetAllDefectas)
	app.Get("/api/sys-defectas/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), defectaHandler.GetDefetaWithItems)
	app.Post("/api/sys-defectas", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), defectaHandler.CreateDefecta)
	app.Put("/api/sys-defectas/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), defectaHandler.UpdateDefecta)
	app.Delete("/api/sys-defectas/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), defectaHandler.DeleteDefecta)
	app.Get("/api/sys-defecta-items", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), defectaHandler.GetAllDefectaItems)
	app.Post("/api/sys-defecta-items", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), defectaHandler.CreateDefectaItem)
	app.Put("/api/sys-defecta-items/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), defectaHandler.UpdateDefectaItem)
	app.Delete("/api/sys-defecta-items/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), defectaHandler.DeleteDefectaItem)

	app.Get("/api/report/neraca-saldo", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), reportHandler.GetNeracaSaldo)
	app.Get("/api/report/profit-by-month", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), reportHandler.GetProfitGraphByMonth)
}
