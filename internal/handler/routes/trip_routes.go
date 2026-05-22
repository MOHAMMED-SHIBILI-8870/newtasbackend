package routes

import (
	"backend/internal/handler"

	"github.com/gofiber/fiber/v2"
)

func TripRoutes(app fiber.Router, h *handler.TripHandler, auth fiber.Handler, admin fiber.Handler) {

	// PUBLIC ROUTES

	publicTrips := app.Group("/trips")
	publicTrips.Get("/", h.GetAllTrips)
	publicTrips.Get("/:name", h.GetTripByName)

	adminTrips := app.Group("/admin/trips", auth, admin)

	adminTrips.Post("/", h.CreateTrip)      // Create master template trip
	adminTrips.Get("/", h.GetAllTrips)      // Admin view all master templates
	adminTrips.Patch("/:id", h.UpdateTrip)  // Admin modifies master details/plans
	adminTrips.Delete("/:id", h.DeleteTrip) // Admin removes a master package
}