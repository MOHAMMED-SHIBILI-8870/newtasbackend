package middleware

import (
	"backend/internal/response"
	"backend/internal/usecase"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func PermissionMiddleware(permissionUsecase *usecase.PermissionUsecase) func(...string) fiber.Handler {
	return func(requiredPermissions ...string) fiber.Handler {
		return func(c *fiber.Ctx) error {
			userID := GetAuthUserID(c)
			if userID == 0 {
				return response.Fail(c, fiber.StatusUnauthorized, "unauthorized", nil)
			}

			role := usecase.NormalizeRole(GetAuthRole(c))
			if role == "admin" {
				return c.Next()
			}

			if len(requiredPermissions) == 0 {
				return c.Next()
			}

			permissions, err := permissionUsecase.GetUserPermissionKeys(c.Context(), userID)
			if err != nil {
				return response.Fail(c, fiber.StatusForbidden, "access denied", err)
			}

			c.Locals(AuthPermissionsKey, permissions)

			permissionSet := make(map[string]struct{}, len(permissions))
			for _, permission := range permissions {
				permissionSet[strings.ToLower(strings.TrimSpace(permission))] = struct{}{}
			}

			for _, requiredPermission := range requiredPermissions {
				if _, ok := permissionSet[strings.ToLower(strings.TrimSpace(requiredPermission))]; ok {
					return c.Next()
				}
			}

			return response.Fail(c, fiber.StatusForbidden, "access denied", nil)
		}
	}
}

func GetAuthPermissions(c *fiber.Ctx) []string {
	raw := c.Locals(AuthPermissionsKey)
	switch v := raw.(type) {
	case []string:
		return v
	case []any:
		permissions := make([]string, 0, len(v))
		for _, item := range v {
			if text, ok := item.(string); ok {
				permissions = append(permissions, text)
			}
		}
		return permissions
	default:
		return []string{}
	}
}
