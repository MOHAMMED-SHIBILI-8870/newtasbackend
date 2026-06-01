package routes

import (
	"backend/internal/handler"

	"github.com/gofiber/fiber/v2"
)

func RBACRoutes(app fiber.Router, roleHandler *handler.RoleHandler, permissionHandler *handler.PermissionHandler, auth fiber.Handler, permission func(...string) fiber.Handler) {
	me := app.Group("/rbac", auth)
	me.Get("/me", roleHandler.GetMyAccess)

	adminRBAC := app.Group("/admin/rbac", auth)
	adminRBAC.Get("/roles", permission("manage_users"), roleHandler.ListRoles)
	adminRBAC.Post("/roles", permission("manage_users"), roleHandler.CreateRole)
	adminRBAC.Patch("/roles/:id", permission("manage_users"), roleHandler.UpdateRole)
	adminRBAC.Delete("/roles/:id", permission("manage_users"), roleHandler.DeleteRole)
	adminRBAC.Patch("/users/:id/role", permission("manage_users"), roleHandler.AssignRoleToUser)

	adminRBAC.Get("/permissions", permission("manage_users"), permissionHandler.ListPermissions)
	adminRBAC.Post("/permissions", permission("manage_users"), permissionHandler.CreatePermission)
	adminRBAC.Patch("/permissions/:id", permission("manage_users"), permissionHandler.UpdatePermission)
	adminRBAC.Delete("/permissions/:id", permission("manage_users"), permissionHandler.DeletePermission)
	adminRBAC.Get("/roles/:id/permissions", permission("manage_users"), permissionHandler.GetRolePermissions)
	adminRBAC.Post("/roles/:id/permissions", permission("manage_users"), permissionHandler.AssignPermissionToRole)
	adminRBAC.Delete("/roles/:id/permissions/:permission_id", permission("manage_users"), permissionHandler.RemovePermissionFromRole)
}
