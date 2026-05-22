package routes

import (
	"backend/internal/handler"

	"github.com/gofiber/fiber/v2"
)

func TripPlanRoutes(app fiber.Router, h *handler.TripPlanHandler, auth fiber.Handler, admin fiber.Handler) {

	adminTrip := app.Group("/admin/plans", auth, admin)

	adminTrip.Post("/trip-plans", h.CreateTripPlan)
	adminTrip.Get("/trip-plans/:trip_id", h.GetTripPlanByID)
	adminTrip.Delete("/trip-plans/:id", h.DeleteTripPlan)
}
