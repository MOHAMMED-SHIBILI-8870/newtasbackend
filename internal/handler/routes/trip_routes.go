package routes

import (
	"backend/internal/handler"
	"backend/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func TripRoutes(app fiber.Router, h *handler.TripHandler) {
	// Create a group and apply your Auth Middleware
	tripGroup := app.Group("/trips", middleware.AuthMiddleware())

	// CRUD Endpoints
	tripGroup.Post("/", h.CreateTrip)             // Create
	tripGroup.Get("/my-trips", h.GetUserTrips)     // Read (List)
	tripGroup.Get("/:id", h.GetTripByID)           // Read (Single)
	tripGroup.Put("/:id", h.UpdateTrip)            // Update
	tripGroup.Delete("/:id", h.DeleteTrip)         // Delete
}