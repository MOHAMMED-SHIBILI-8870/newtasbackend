package routes

import (
	"backend/internal/handler"
	"backend/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func TripPlanRoutes(app fiber.Router, h *handler.TripPlanHandler) {

	adminTrip := app.Group("/admin/plans", middleware.AuthMiddleware(), middleware.RoleMiddleware("admin"))

	adminTrip.Post("/trip-plans", h.CreateTripPlan)
	adminTrip.Get("/trip-plans/:trip_id", h.GetTripPlanByID)
	adminTrip.Delete("/trip-plans/:id", h.DeleteTripPlan)
}
