package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "github.com/heru-oktafian/fiber-apotek/controllers/transactions"
	middlewares "github.com/heru-oktafian/fiber-apotek/middlewares"
)

func TransDuplicateReceiptRoutes(app *fiber.App) {
	// Routes untuk duplicate receipt
	app.Post("/api/duplicate-receipts", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.CreateDuplicateReceipt)
	app.Get("/api/duplicate-receipts", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.GetAllDuplicateReceipts)
	app.Get("/api/duplicate-receipts/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.GetDuplicateWithItems)
	app.Put("/api/duplicate-receipts/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.UpdateDuplicateReceipt)
	app.Delete("/api/duplicate-receipts/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.DeleteDuplicateReceipt)

	// Routes untuk duplicate receipt items
	app.Get("/api/duplicate-receipts-items/all/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.GetAllDuplicateItems)
	app.Post("/api/duplicate-receipts-items", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.CreateDuplicateReceiptItem)
	app.Put("/api/duplicate-receipts-items/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.UpdateDuplicateReceiptItem)
	app.Delete("/api/duplicate-receipts-items/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.DeleteDuplicateReceiptItem)

	// Routes untuk duplicate receipt details
	app.Get("/api/duplicate-receipts-details", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), controllers.GetAllDuplicateDetail)
}
