package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "github.com/heru-oktafian/fiber-apotek/controllers/systems"
	middlewares "github.com/heru-oktafian/fiber-apotek/middlewares"
)

func SysDailyAssetRoute(app *fiber.App) {
	app.Get("/api/daily_asset", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), controllers.GetAllAssets)
}
