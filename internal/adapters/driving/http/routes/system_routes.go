package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	legacyroutes "apotek-clean/routes"
)

func RegisterSystemRoutes(app *fiber.App) {
	legacyroutes.SysDashboardRoute(app)
	legacyroutes.SysDailyAssetRoute(app)
	legacyroutes.SysDefectaRoute(app)
	legacyroutes.SysReportRoute(app)
	legacyroutes.SysUserBranchRoutes(app)
}
