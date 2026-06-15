package middleware

import (
	"backend/internal/repository"
	"backend/internal/response"
	"backend/internal/usecase"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func extractToken(c *fiber.Ctx) (string, error) {
	authHeader := c.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer "), nil
	}

	if token := c.Cookies("access_token"); token != "" {
		return token, nil
	}

	if token := c.Query("token"); token != "" {
		return token, nil
	}

	return "", fiber.NewError(fiber.StatusUnauthorized, "missing authorization token")
}

// AuthMiddleware validates JWT, checks user state, and sets auth context.
func AuthMiddleware(userRepo repository.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, err := extractToken(c)
		if err != nil {
			return response.Fail(c, fiber.StatusUnauthorized, "invalid token", err)
		}

		userID, role, err := usecase.ValidateJwt(token)
		if err != nil {
			return response.Fail(c, fiber.StatusUnauthorized, "invalid token", err)
		}

		user, err := userRepo.GetByID(c.Context(), userID)
		if err != nil {
			return response.Fail(c, fiber.StatusUnauthorized, "invalid token", err)
		}

		if user == nil {
			return response.Fail(c, fiber.StatusUnauthorized, "invalid token", nil)
		}

		if user.IsBlocked {
			return response.Fail(c, fiber.StatusForbidden, "account is blocked", nil)
		}

		if !user.IsVerified {
			return response.Fail(c, fiber.StatusForbidden, "account is not verified", nil)
		}

		currentRole := usecase.NormalizeRole(user.Role)
		if currentRole != role {
			return response.Fail(c, fiber.StatusUnauthorized, "token role mismatch", nil)
		}

		c.Locals(AuthUserIDKey, user.ID)
		c.Locals(AuthRoleKey, currentRole)
		c.Locals(AuthEmailKey, user.Email)

		return c.Next()
	}
}

func RoleMiddleware(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals(AuthRoleKey).(string)
		if !ok {
			return response.Fail(c, fiber.StatusForbidden, "role not found", nil)
		}

		role = usecase.NormalizeRole(role)
		for _, allowed := range allowedRoles {
			if role == usecase.NormalizeRole(allowed) {
				return c.Next()
			}
		}

		return response.Fail(c, fiber.StatusForbidden, "access denied", nil)
	}
}

func GetAuthUserID(c *fiber.Ctx) uint {
	if v, ok := c.Locals(AuthUserIDKey).(uint); ok {
		return v
	}
	switch v := c.Locals(AuthUserIDKey).(type) {
	case int:
		return uint(v)
	case int64:
		return uint(v)
	case float64:
		return uint(v)
	case string:
		id, _ := strconv.Atoi(v)
		return uint(id)
	default:
		return 0
	}
}

func GetAuthRole(c *fiber.Ctx) string {
	if v, ok := c.Locals(AuthRoleKey).(string); ok {
		return usecase.NormalizeRole(v)
	}
	return ""
}

func GetAuthEmail(c *fiber.Ctx) string {
	if v, ok := c.Locals(AuthEmailKey).(string); ok {
		return v
	}
	return ""
}
