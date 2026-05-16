package routes

import (
    "backend/internal/handler"
    "backend/internal/middleware"
    "github.com/gofiber/fiber/v2"
)

func AdminRoutes(app *fiber.App, h *handler.AdminHandler) {
    // We create a group so all these routes start with /api/v1/admin (or just /admin)
    // The middlewares here protect EVERY route inside this group
    adminGroup := app.Group("/admin", middleware.AuthMiddleware(), middleware.RoleMiddleware("admin"))

    // 1. List & Search Users
    // Use ?role=guide or ?search=name to filter
    adminGroup.Get("/users", h.ListUsers)

    // 2. Block/Unblock User
    // This toggles the 'is_blocked' status in the DB
    adminGroup.Patch("/users/:id/block", h.BlockUser)

    // 3. Change User Role
    // This allows promoting a 'user' to a 'guide' or 'manager'
    adminGroup.Patch("/users/:id/role", h.UpdateRole)

    adminGroup.Post("/createUser",h.CreateUserByAdmin)

    adminGroup.Delete("/remove/:id",h.DeleteUser)
}