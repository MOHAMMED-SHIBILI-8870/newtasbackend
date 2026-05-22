package routes

import (
	"backend/internal/handler"

	"github.com/gofiber/fiber/v2"
)

func NotificationRoutes(app fiber.Router, h *handler.NotificationHandler, auth fiber.Handler, admin fiber.Handler) {
	notifications := app.Group("/notifications", auth)
	notifications.Get("/", h.GetNotifications)
	notifications.Patch("/:id/read", h.MarkAsRead)

	adminNotifications := app.Group("/admin/notifications", auth, admin)
	adminNotifications.Get("/", h.GetNotifications)
	adminNotifications.Patch("/:id/read", h.MarkAsRead)
}
