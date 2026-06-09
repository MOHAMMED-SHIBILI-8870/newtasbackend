package routes

import (
	"backend/internal/handler"

	"github.com/gofiber/fiber/v2"
)

func TripSlotRoutes(app fiber.Router, h *handler.TripSlotHandler, auth fiber.Handler, admin fiber.Handler) {
	publicTrips := app.Group("/trips")
	publicTrips.Get("/:trip_id/slots", h.GetUpcomingSlotsByTripID)

	adminSlots := app.Group("/admin/slots", auth, admin)
	adminSlots.Post("/", h.CreateSlot)
	adminSlots.Get("/", h.ListSlots)
	adminSlots.Get("/:id", h.GetSlotByID)
	adminSlots.Put("/:id", h.UpdateSlot)
	adminSlots.Delete("/:id", h.DeleteSlot)
}
