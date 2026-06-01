package routes

import (
	"backend/internal/handler"

	"github.com/gofiber/fiber/v2"
)

func TrackingRoutes(app fiber.Router, h *handler.TrackingHandler, auth fiber.Handler, permission func(...string) fiber.Handler) {
	tracking := app.Group("/tracking", auth)
	tracking.Post("/", permission("manage_tracking"), h.UpdateLocation)
	tracking.Get("/booking/:booking_id", h.GetLatestByBookingID)
	tracking.Get("/booking/:booking_id/history", h.GetTrackingsByBookingID)

	adminTracking := app.Group("/admin/tracking", auth)
	adminTracking.Get("/", permission("manage_tracking"), h.GetAllTracking)
}
