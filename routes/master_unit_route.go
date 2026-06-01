package routes

import (
	fiber "github.com/gofiber/fiber/v2"
	controllers "github.com/heru-oktafian/fiber-apotek/controllers/masters"
	middlewares "github.com/heru-oktafian/fiber-apotek/middlewares"
)

func MasterUnitRoutes(app *fiber.App) {
	app.Get("/api/units", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.GetAllUnit)
	app.Get("/api/units/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.GetUnit)
	app.Post("/api/units", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.CreateUnit)
	app.Put("/api/units/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.UpdateUnit)
	app.Delete("/api/units/:id", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.DeleteUnit)
	app.Get("/api/cmb-units", middlewares.JWTMiddleware, middlewares.RoleMiddleware("administrator", "superadmin"), controllers.CmbUnit)
}
