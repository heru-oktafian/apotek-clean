package routes

import (
	handlers "apotek-clean/internal/adapters/driving/http/handlers"
	middlewares "apotek-clean/middlewares"
	fiber "github.com/gofiber/fiber/v2"
)

func RegisterMasterDataRoutes(app *fiber.App) {
	productCategoryHandler := handlers.NewProductCategoryHandler()
	productHandler := handlers.NewProductHandler()
	supplierCategoryHandler := handlers.NewSupplierCategoryHandler()
	supplierHandler := handlers.NewSupplierHandler()
	unitHandler := handlers.NewUnitHandler()
	unitConversionHandler := handlers.NewUnitConversionHandler()
	memberCategoryHandler := handlers.NewMemberCategoryHandler()
	memberHandler := handlers.NewMemberHandler()

	app.Get("/api/product-categories", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), productCategoryHandler.GetAllProductCategory)
	app.Post("/api/product-categories", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), productCategoryHandler.CreateProductCategory)
	app.Get("/api/product-categories/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), productCategoryHandler.GetProductCategory)
	app.Put("/api/product-categories/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), productCategoryHandler.UpdateProductCategory)
	app.Delete("/api/product-categories/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), productCategoryHandler.DeleteProductCategory)
	app.Get("/api/product-categories-combo", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), productCategoryHandler.CmbProductCategory)

	app.Get("/api/products", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), productHandler.GetAllProduct)
	app.Post("/api/products", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), productHandler.CreateProduct)
	app.Get("/api/products/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), productHandler.GetProduct)
	app.Put("/api/products/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), productHandler.UpdateProduct)
	app.Delete("/api/products/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), productHandler.DeleteProduct)
	app.Get("/api/sales-products-combo", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), productHandler.CmbProdSale)
	app.Get("/api/sale-products-combo", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), productHandler.CmbProdSale)
	app.Get("/api/purchase-products-combo", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), productHandler.CmbProdPurchase)

	app.Get("/api/supplier-categories", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), supplierCategoryHandler.GetAllSupplierCategory)
	app.Post("/api/supplier-categories", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), supplierCategoryHandler.CreateSupplierCategory)
	app.Get("/api/supplier-categories/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), supplierCategoryHandler.GetSupplierCategoryByID)
	app.Put("/api/supplier-categories/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), supplierCategoryHandler.UpdateSupplierCategory)
	app.Delete("/api/supplier-categories/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), supplierCategoryHandler.DeleteSupplierCategory)
	app.Get("/api/supplier-categories-combo", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), supplierCategoryHandler.CmbSupplierCategory)

	app.Get("/api/suppliers", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), supplierHandler.GetAllSupplier)
	app.Post("/api/suppliers", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), supplierHandler.CreateSupplier)
	app.Get("/api/suppliers/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), supplierHandler.GetSupplierByID)
	app.Put("/api/suppliers/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), supplierHandler.UpdateSupplier)
	app.Delete("/api/suppliers/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), supplierHandler.DeleteSupplier)
	app.Get("/api/suppliers-combo", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), supplierHandler.CmbSupplier)

	app.Get("/api/units", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin", "operator", "finance", "cashier"), unitHandler.GetAllUnit)
	app.Get("/api/units/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin", "operator", "finance", "cashier"), unitHandler.GetUnit)
	app.Post("/api/units", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), unitHandler.CreateUnit)
	app.Put("/api/units/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), unitHandler.UpdateUnit)
	app.Delete("/api/units/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), unitHandler.DeleteUnit)
	app.Get("/api/units-combo", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin", "operator", "finance", "cashier"), unitHandler.CmbUnit)

	app.Get("/api/unit-conversions", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin", "operator", "finance", "cashier"), unitConversionHandler.GetAllUnitConversion)
	app.Get("/api/unit-conversions/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin", "operator", "finance", "cashier"), unitConversionHandler.GetUnitConversionByID)
	app.Post("/api/unit-conversions", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), unitConversionHandler.CreateUnitConversion)
	app.Put("/api/unit-conversions/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), unitConversionHandler.UpdateUnitConversion)
	app.Delete("/api/unit-conversions/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), unitConversionHandler.DeleteUnitConversion)
	app.Get("/api/conversion-products-combo", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin", "operator", "finance", "cashier"), unitConversionHandler.CmbProdConv)

	app.Get("/api/member-categories", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), memberCategoryHandler.GetAllMemberCategory)
	app.Get("/api/member-categories/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), memberCategoryHandler.GetMemberCategory)
	app.Post("/api/member-categories", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), memberCategoryHandler.CreateMemberCategory)
	app.Put("/api/member-categories/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), memberCategoryHandler.UpdateMemberCategory)
	app.Delete("/api/member-categories/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), memberCategoryHandler.DeleteMemberCategory)
	app.Get("/api/member-categories-combo", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), memberCategoryHandler.CmbMemberCategory)

	app.Get("/api/members", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), memberHandler.GetAllMember)
	app.Get("/api/members/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), memberHandler.GetMember)
	app.Post("/api/members", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), memberHandler.CreateMember)
	app.Put("/api/members/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), memberHandler.UpdateMember)
	app.Delete("/api/members/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "finance", "superadmin"), memberHandler.DeleteMember)
	app.Get("/api/members-combo", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "operator", "cashier", "finance", "superadmin"), memberHandler.CmbMember)
}
