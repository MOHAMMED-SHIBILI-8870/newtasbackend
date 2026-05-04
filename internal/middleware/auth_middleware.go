package middleware

import (
	"backend/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing Authorization header",
			})
		}

		// Expect: "Bearer <token>"
		const prefix = "Bearer "
		if len(authHeader) <= len(prefix) || authHeader[:len(prefix)] != prefix {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid Authorization format",
			})
		}

		accessToken := authHeader[len(prefix):]

		// Validate token
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

		c.Locals("user_id", uint(userID))
		c.Locals("role", role)

		return c.Next()
	}
}


func RoleMiddleware(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {

		role, ok := c.Locals("role").(string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "role not found",
			})
		}

		for _, allowed := range allowedRoles {
			if role == allowed {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied",
		})
	}
}