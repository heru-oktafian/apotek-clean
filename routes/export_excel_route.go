package routes

import (
	"log"

	fiber "github.com/gofiber/fiber/v2"
	configs "github.com/heru-oktafian/fiber-apotek/configs"
	middlewares "github.com/heru-oktafian/fiber-apotek/middlewares"
	export_services "github.com/heru-oktafian/fiber-apotek/services/exports"
	excel_audits "github.com/heru-oktafian/fiber-apotek/services/exports/excel/audits"
	excel_masters "github.com/heru-oktafian/fiber-apotek/services/exports/excel/masters"
	excel_reports "github.com/heru-oktafian/fiber-apotek/services/exports/excel/reports"
	excel_systems "github.com/heru-oktafian/fiber-apotek/services/exports/excel/systems"
	excel_transactions "github.com/heru-oktafian/fiber-apotek/services/exports/excel/transactions"
)

func ExportExcelRoutes(app *fiber.App) {
	log.Println("[Route] Initializing ExportExcelRoutes")

	// Inisialisasi Excel Service dan semua Handlers
	excelService := export_services.NewExcelServices(configs.DB)
	excelProductHandler := excel_masters.NewExcelProductHandler(excelService)
	excelUnitHandler := excel_masters.NewExcelUnitHandler(excelService)
	excelProductCategoryHandler := excel_masters.NewExcelProductCategoryHandler(excelService)
	excelUnitConversionHandler := excel_masters.NewExcelUnitConversionHandler(excelService)
	excelSupplierHandler := excel_masters.NewExcelSupplierHandler(excelService)
	excelSupplierCategoryHandler := excel_masters.NewExcelSupplierCategoryHandler(excelService)
	excelMemberCategoryHandler := excel_masters.NewExcelMemberCategoryHandler(excelService)
	excelMemberHandler := excel_masters.NewExcelMemberHandler(excelService)
	excelFirstStockHandler := excel_audits.NewExcelFirstStockHandler(excelService)
	excelOpnameHandler := excel_audits.NewExcelOpnameHandler(excelService)
	excelFirstStockItemHandler := excel_audits.NewExcelFirstStockItemHandler(excelService)
	excelOpnameItemHandler := excel_audits.NewExcelOpnameItemHandler(excelService)

	// Transaction Handlers
	excelPurchaseHandler := excel_transactions.NewExcelPurchaseHandler(excelService)
	excelSaleHandler := excel_transactions.NewExcelSaleHandler(excelService)
	excelDuplicateReceiptHandler := excel_transactions.NewExcelDuplicateReceiptHandler(excelService)
	excelBuyReturnHandler := excel_transactions.NewExcelBuyReturnHandler(excelService)
	excelSaleReturnHandler := excel_transactions.NewExcelSaleReturnHandler(excelService)
	excelExpenseHandler := excel_transactions.NewExcelExpenseHandler(excelService)
	excelAnotherIncomeHandler := excel_transactions.NewExcelAnotherIncomeHandler(excelService)
	excelPurchaseItemHandler := excel_transactions.NewExcelPurchaseItemHandler(excelService)
	excelSaleItemHandler := excel_transactions.NewExcelSaleItemHandler(excelService)
	excelDuplicateReceiptItemHandler := excel_transactions.NewExcelDuplicateReceiptItemHandler(excelService)
	excelBuyReturnItemHandler := excel_transactions.NewExcelBuyReturnItemHandler(excelService)
	excelSaleReturnItemHandler := excel_transactions.NewExcelSaleReturnItemHandler(excelService)

	// Systems Handlers
	excelDailyAssetHandler := excel_systems.NewExcelDailyAssetHandler(excelService)
	excelDefectaHandler := excel_systems.NewExcelDefectaHandler(excelService)
	excelNearedReportHandler := excel_systems.NewExcelNearedReportHandler(excelService)
	excelTopSellingHandler := excel_systems.NewExcelTopSellingHandler(excelService)
	excelLeastSellingHandler := excel_systems.NewExcelLeastSellingHandler(excelService)

	// Reports Handlers
	excelNeracaSaldoHandler := excel_reports.NewExcelNeracaSaldoHandler(excelService)

	// Protected routes: Export ke Excel
	app.Get("/api/products/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelProductHandler.ExportExcel)
	app.Get("/api/units/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelUnitHandler.ExportExcel)
	app.Get("/api/product-categories/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelProductCategoryHandler.ExportExcel)
	app.Get("/api/unit-conversions/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelUnitConversionHandler.ExportExcel)
	app.Get("/api/suppliers/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelSupplierHandler.ExportExcel)
	app.Get("/api/supplier-categories/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelSupplierCategoryHandler.ExportExcel)
	app.Get("/api/member-categories/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelMemberCategoryHandler.ExportExcel)
	app.Get("/api/members/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelMemberHandler.ExportExcel)
	app.Get("/api/first-stocks/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelFirstStockHandler.ExportExcel)
	app.Get("/api/opnames/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelOpnameHandler.ExportExcel)
	app.Get("/api/first-stock-items/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelFirstStockItemHandler.ExportExcel)
	app.Get("/api/opname-items/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelOpnameItemHandler.ExportExcel)

	// Transaction Routes
	app.Get("/api/purchases/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelPurchaseHandler.ExportExcel)
	app.Get("/api/sales/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelSaleHandler.ExportExcel)
	app.Get("/api/duplicate-receipts/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelDuplicateReceiptHandler.ExportExcel)
	app.Get("/api/buy-returns/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelBuyReturnHandler.ExportExcel)
	app.Get("/api/sale-returns/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelSaleReturnHandler.ExportExcel)
	app.Get("/api/expenses/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelExpenseHandler.ExportExcel)
	app.Get("/api/another-incomes/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelAnotherIncomeHandler.ExportExcel)
	app.Get("/api/purchase-items/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelPurchaseItemHandler.ExportExcel)
	app.Get("/api/sale-items/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelSaleItemHandler.ExportExcel)
	app.Get("/api/duplicate-receipt-items/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelDuplicateReceiptItemHandler.ExportExcel)
	app.Get("/api/buy-return-items/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelBuyReturnItemHandler.ExportExcel)
	app.Get("/api/sale-return-items/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelSaleReturnItemHandler.ExportExcel)

	// Systems Routes
	app.Get("/api/daily-assets/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelDailyAssetHandler.ExportExcel)
	app.Get("/api/defectas/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelDefectaHandler.ExportExcel)
	app.Get("/api/dashboard/neared-report/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelNearedReportHandler.ExportExcel)
	app.Get("/api/dashboard/top-selling-report/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelTopSellingHandler.ExportExcel)
	app.Get("/api/dashboard/least-selling-report/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelLeastSellingHandler.ExportExcel)

	// Reports Routes
	app.Get("/api/reports/neraca-saldo/excel", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), excelNeracaSaldoHandler.ExportExcel)

	log.Println("[Route] ExportExcelRoutes initialized successfully!")
}
