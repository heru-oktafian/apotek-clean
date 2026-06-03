package routes

import fiber "github.com/gofiber/fiber/v2"

func RegisterCoreRoutes(app *fiber.App) {
	RegisterAuthRoutes(app)
	RegisterUserRoutes(app)
	RegisterBranchRoutes(app)
	RegisterExportRoutes(app)
	RegisterMasterDataRoutes(app)
	RegisterTransactionRoutes(app)
	RegisterAuditRoutes(app)
	RegisterSystemRoutes(app)
}
