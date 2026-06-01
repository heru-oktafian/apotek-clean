package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	legacyroutes "apotek-clean/routes"
)

func RegisterTransactionRoutes(app *fiber.App) {
	legacyroutes.TransAnotherIncomeRoute(app)
	legacyroutes.TransBuyReturnRoutes(app)
	legacyroutes.TransDuplicateReceiptRoutes(app)
	legacyroutes.TransExpenseRoutes(app)
	legacyroutes.TransPurchaseRoutes(app)
	legacyroutes.TransSaleRoutes(app)
	legacyroutes.TransSaleReturnRoutes(app)
}
