package middleware

import (
	"backend/internal/usecase"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware validates JWT and sets user info in context
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing Authorization header",
			})
		}

		// Check Bearer format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid Authorization format",
			})
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate JWT
		userID, role, err := usecase.ValidateJwt(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		// Store values in context
		c.Locals("user_id", uint(userID))
		c.Locals("role", role)

		return c.Next()
	}
}
func RoleMiddleware(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Get role from context
		role, ok := c.Locals("role").(string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "role not found",
			})
		}

		// Check if role is allowed
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