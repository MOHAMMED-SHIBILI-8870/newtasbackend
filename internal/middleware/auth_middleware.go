package middleware

import (
	"backend/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {

		accessToken := c.Cookies("access_token")
		if accessToken == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authentication required",
			})
		}

		userID, role, err := usecase.ValidateJwt(accessToken)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token",
			})
		}

		if role == "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "admins are not allowed",
			})
		}

		// Set values in context
		c.Locals("userID", userID)
		c.Locals("role", role)

		return c.Next()
	}
}

func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")

		if role != "admin" {
			return c.Status(403).JSON(fiber.Map{
				"error": "admin only access",
			})
		}

		return c.Next()
	}
}