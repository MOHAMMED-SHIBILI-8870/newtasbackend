package routes

import (
	"backend/internal/handler"
	"github.com/gofiber/fiber/v2"
)

func AdminRoutes(app *fiber.App, h *handler.AdminHandler, auth fiber.Handler, admin fiber.Handler) {
	adminGroup := app.Group("/admin", auth, admin)

	// 1. List & Search Users
	// Use ?role=guide or ?search=name to filter
	adminGroup.Get("/users", h.ListUsers)

	adminGroup.Patch("/users/:id/block", h.BlockUser)

	adminGroup.Patch("/users/:id/role", h.UpdateRole)

	adminGroup.Post("/users", h.CreateUserByAdmin)

	adminGroup.Delete("/users/:id", h.DeleteUser)
}
