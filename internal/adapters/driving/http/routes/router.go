package routes

import fiber "github.com/gofiber/fiber/v2"

func RegisterCoreRoutes(app *fiber.App) {
	RegisterAuthRoutes(app)
	RegisterUserRoutes(app)
	RegisterBranchRoutes(app)
	RegisterMasterDataRoutes(app)
	RegisterTransactionRoutes(app)
	RegisterAuditRoutes(app)
	RegisterSystemRoutes(app)
	RegisterExportRoutes(app)
}
