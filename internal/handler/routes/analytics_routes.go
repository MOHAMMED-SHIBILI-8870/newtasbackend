package routes

import (
	"backend/internal/handler"

	"github.com/gofiber/fiber/v2"
)

func AnalyticsRoutes(app fiber.Router, h *handler.AnalyticsHandler, auth fiber.Handler, permission func(...string) fiber.Handler) {
	adminGroup := app.Group("/admin/analytics", auth)
	adminGroup.Get("/dashboard", permission("view_analytics"), h.GetDashboard)
}
