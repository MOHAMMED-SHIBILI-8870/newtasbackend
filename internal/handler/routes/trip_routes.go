package routes

import (
	"backend/internal/handler"
	"backend/internal/middleware"

	"github.com/gofiber/fiber/v2"
)


func TripRoutes(app fiber.Router, h *handler.TripHandler) {

	// Logged-in users + admin can view
	trip := app.Group("/trips", middleware.AuthMiddleware())

	trip.Get("/my-trips", h.GetUserTrips)
	trip.Get("/", h.GetAllTrips)
	trip.Get("/:id", h.GetTripByID)

	// Admin only
	adminTrip := app.Group("/admin/trips",middleware.AuthMiddleware(),middleware.RoleMiddleware("admin"))

	adminTrip.Post("/", h.CreateTrip)
	adminTrip.Patch("/:id", h.UpdateTrip)
	adminTrip.Delete("/:id", h.DeleteTrip)
}