package middlewares

import (
	strings "strings"

	fiber "github.com/gofiber/fiber/v2"
	helpers "github.com/heru-oktafian/fiber-apotek/helpers"
	models "github.com/heru-oktafian/fiber-apotek/models"
	services "github.com/heru-oktafian/fiber-apotek/services"
)

func JWTMiddleware(c *fiber.Ctx) error {
	return helpers.TokenValidation(c, "sub")
}

func RoleMiddleware(allowedRoles ...models.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, _ := services.GetUserRole(c)
		userRole = strings.ToLower(userRole)

		for _, role := range allowedRoles {
			if strings.ToLower(string(role)) == userRole {
				return c.Next()
			}
		}

		return helpers.JSONResponse(c, fiber.StatusForbidden, "Forbidden", "You don't have permission to access this resource!")
	}
}
