package routes

import (
	"log"

	fiber "github.com/gofiber/fiber/v2"
	configs "github.com/heru-oktafian/fiber-apotek/configs"
	middlewares "github.com/heru-oktafian/fiber-apotek/middlewares"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
	pdf_audits "github.com/heru-oktafian/fiber-apotek/services/exports/pdf/audits"
	pdf_masters "github.com/heru-oktafian/fiber-apotek/services/exports/pdf/masters"
	pdf_transactions "github.com/heru-oktafian/fiber-apotek/services/exports/pdf/transactions"
)

func ExportPDFRoutes(app *fiber.App) {
	log.Println("[Route] Initializing ExportPDFRoutes")

	// Inisialisasi PDF Service
	pdfService := export_services.NewPDFService(configs.DB)

	// Inisialisasi handlers untuk semua entitas
	pdfProductHandler := pdf_masters.NewPdfProductHandler(pdfService)
	pdfUnitHandler := pdf_masters.NewPdfUnitHandler(pdfService)
	pdfProductCategoryHandler := pdf_masters.NewPdfProductCategoryHandler(pdfService)
	pdfUnitConversionHandler := pdf_masters.NewPdfUnitConversionHandler(pdfService)
	pdfSupplierHandler := pdf_masters.NewPdfSupplierHandler(pdfService)
	pdfSupplierCategoryHandler := pdf_masters.NewPdfSupplierCategoryHandler(pdfService)
	pdfMemberCategoryHandler := pdf_masters.NewPdfMemberCategoryHandler(pdfService)
	pdfMemberHandler := pdf_masters.NewPdfMemberHandler(pdfService)
	pdfFirstStockHandler := pdf_audits.NewPdfFirstStockHandler(pdfService)
	pdfOpnameHandler := pdf_audits.NewPdfOpnameHandler(pdfService)
	pdfFirstStockItemHandler := pdf_audits.NewPdfFirstStockItemHandler(pdfService)
	pdfOpnameItemHandler := pdf_audits.NewPdfOpnameItemHandler(pdfService)
	pdfPurchaseItemHandler := pdf_transactions.NewPdfPurchaseItemHandler(pdfService)
	pdfSaleItemHandler := pdf_transactions.NewPdfSaleItemHandler(pdfService)
	pdfBuyReturnItemHandler := pdf_transactions.NewPdfBuyReturnItemHandler(pdfService)
	pdfSaleReturnItemHandler := pdf_transactions.NewPdfSaleReturnItemHandler(pdfService)
	pdfDuplicateReceiptItemHandler := pdf_transactions.NewPdfDuplicateReceiptItemHandler(pdfService)

	// Handler untuk list/summary PDF
	pdfPurchaseHandler := pdf_transactions.NewPdfPurchaseHandler(pdfService)
	pdfSaleHandler := pdf_transactions.NewPdfSaleHandler(pdfService)
	pdfBuyReturnHandler := pdf_transactions.NewPdfBuyReturnHandler(pdfService)
	pdfSaleReturnHandler := pdf_transactions.NewPdfSaleReturnHandler(pdfService)
	pdfDuplicateReceiptHandler := pdf_transactions.NewPdfDuplicateReceiptHandler(pdfService)
	pdfExpenseHandler := pdf_transactions.NewPdfExpenseHandler(pdfService)
	pdfAnotherIncomeHandler := pdf_transactions.NewPdfAnotherIncomeHandler(pdfService)
	pdfProductLabelHandler := pdf_masters.NewPdfProductLabelHandler(pdfService)

	// Protected routes untuk export PDF
	app.Get("/api/product-label/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfProductLabelHandler.ExportPDF)
	app.Get("/api/products/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfProductHandler.ExportPDF)
	app.Get("/api/units/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfUnitHandler.ExportPDF)
	app.Get("/api/product-categories/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfProductCategoryHandler.ExportPDF)
	app.Get("/api/unit-conversions/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfUnitConversionHandler.ExportPDF)
	app.Get("/api/suppliers/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfSupplierHandler.ExportPDF)
	app.Get("/api/supplier-categories/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfSupplierCategoryHandler.ExportPDF)
	app.Get("/api/member-categories/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfMemberCategoryHandler.ExportPDF)
	app.Get("/api/members/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfMemberHandler.ExportPDF)
	app.Get("/api/first-stocks/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfFirstStockHandler.ExportPDF)
	app.Get("/api/opnames/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfOpnameHandler.ExportPDF)
	app.Get("/api/first-stock-items/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfFirstStockItemHandler.ExportPDF)
	app.Get("/api/opname-items/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfOpnameItemHandler.ExportPDF)
	app.Get("/api/purchase-items/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfPurchaseItemHandler.ExportPDF)
	app.Get("/api/sale-items/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfSaleItemHandler.ExportPDF)
	app.Get("/api/buy-return-items/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfBuyReturnItemHandler.ExportPDF)
	app.Get("/api/sale-return-items/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfSaleReturnItemHandler.ExportPDF)
	app.Get("/api/duplicate-receipt-items/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfDuplicateReceiptItemHandler.ExportPDF)

	// Routes untuk list/summary PDF dengan filter month
	app.Get("/api/purchases/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfPurchaseHandler.ExportPDF)
	app.Get("/api/sales/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfSaleHandler.ExportPDF)
	app.Get("/api/duplicate-receipts/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfDuplicateReceiptHandler.ExportPDF)
	app.Get("/api/buy-returns/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfBuyReturnHandler.ExportPDF)
	app.Get("/api/sale-returns/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfSaleReturnHandler.ExportPDF)
	app.Get("/api/expenses/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfExpenseHandler.ExportPDF)
	app.Get("/api/another-incomes/pdf", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), pdfAnotherIncomeHandler.ExportPDF)

	log.Println("[Route] ExportPDFRoutes initialized successfully!")
}
