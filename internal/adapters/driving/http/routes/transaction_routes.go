package routes

import (
	handlers "apotek-clean/internal/adapters/driving/http/handlers"
	middlewares "apotek-clean/middlewares"
	fiber "github.com/gofiber/fiber/v2"
)

func RegisterTransactionRoutes(app *fiber.App) {
	app.Get("/api/purchases", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.GetAllPurchases)
	app.Get("/api/purchases/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.GetPurchaseWithItems)
	app.Post("/api/purchases", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.CreatePurchaseTransaction)
	app.Put("/api/purchases/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.UpdatePurchase)
	app.Delete("/api/purchases/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.DeletePurchase)
	app.Get("/api/purchase-items/all/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.GetAllPurchaseItems)
	app.Post("/api/purchase-items", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.CreatePurchaseItem)
	app.Put("/api/purchase-items/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.UpdatePurchaseItem)
	app.Delete("/api/purchase-items/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.DeletePurchaseItem)

	app.Get("/api/sales", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.GetAllSales)
	app.Get("/api/sales/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.GetSaleWithItems)
	app.Post("/api/sales", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.CreateSaleTransaction)
	app.Put("/api/sales/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.UpdateSale)
	app.Delete("/api/sales/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.DeleteSale)
	app.Get("/api/sale-items/all/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.GetAllSaleItems)
	app.Post("/api/sale-items", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.CreateSaleItem)
	app.Put("/api/sale-items/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.UpdateSaleItem)
	app.Delete("/api/sale-items/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.DeleteSaleItem)
	app.Get("/api/sales-details", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.GetAllSalesDetail)

	app.Get("/api/another-incomes", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.GetAllAnotherIncomes)
	app.Post("/api/another-incomes", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), handlers.CreateAnotherIncome)
	app.Post("/api/another-incomes/", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), handlers.CreateAnotherIncome)
	app.Put("/api/another-incomes/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), handlers.UpdateAnotherIncome)
	app.Delete("/api/another-incomes/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), handlers.DeleteAnotherIncome)

	app.Get("/api/buy-returns", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.GetAllBuyReturns)
	app.Post("/api/buy-returns", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.CreateBuyReturnTransaction)
	app.Get("/api/buy-returns/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.GetBuyReturnWithItems)
	app.Get("/api/cmb-prod-buy-returns", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.GetBuyItemsForReturn)
	app.Get("/api/cmb-purchases", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), handlers.CmbPurchase)

	app.Post("/api/duplicate-receipts", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.CreateDuplicateReceipt)
	app.Get("/api/duplicate-receipts", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.GetAllDuplicateReceipts)
	app.Get("/api/duplicate-receipts/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.GetDuplicateWithItems)
	app.Put("/api/duplicate-receipts/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.UpdateDuplicateReceipt)
	app.Delete("/api/duplicate-receipts/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.DeleteDuplicateReceipt)
	app.Get("/api/duplicate-receipts-items/all/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.GetAllDuplicateItems)
	app.Post("/api/duplicate-receipts-items", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.CreateDuplicateReceiptItem)
	app.Put("/api/duplicate-receipts-items/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.UpdateDuplicateReceiptItem)
	app.Delete("/api/duplicate-receipts-items/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.DeleteDuplicateReceiptItem)
	app.Get("/api/duplicate-receipts-details", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.GetAllDuplicateDetail)

	app.Post("/api/expenses", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.CreateExpense)
	app.Get("/api/expenses", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.GetAllExpenses)
	app.Put("/api/expenses/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.UpdateExpense)
	app.Delete("/api/expenses/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.DeleteExpense)

	app.Get("/api/sale-returns", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.GetAllSaleReturns)
	app.Get("/api/sale-returns/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.GetSaleReturnWithItems)
	app.Post("/api/sale-returns", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.CreateSaleReturnTransaction)
	app.Get("/api/cmb-prod-sale-returns", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.GetSaleItemsForReturn)
	app.Get("/api/cmb-sales", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), handlers.CmbSale)
}
