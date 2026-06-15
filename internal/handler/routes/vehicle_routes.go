package routes

import (
	"backend/internal/handler"

	"github.com/gofiber/fiber/v2"
)

func VehicleRoutes(app fiber.Router, h *handler.VehicleHandler, auth fiber.Handler, permission func(...string) fiber.Handler) {
	vehicles := app.Group("/vehicles", auth)
	vehicles.Get("/", h.ListVehicles)
	vehicles.Get("/trip/:trip_id", h.GetVehicleByTripID)
	vehicles.Get("/:id", h.GetVehicleByID)

	adminVehicles := app.Group("/admin/vehicles", auth)
	adminVehicles.Get("/", permission("manage_fleet"), h.ListVehicles)
	adminVehicles.Post("/", permission("manage_fleet"), h.CreateVehicle)
	adminVehicles.Put("/:id", permission("manage_fleet"), h.UpdateVehicle)
	adminVehicles.Delete("/:id", permission("manage_fleet"), h.DeleteVehicle)
	adminVehicles.Patch("/:id/assign-trip", permission("manage_fleet"), h.AssignVehicleToTrip)
	adminVehicles.Patch("/:id/assign-driver", permission("manage_fleet"), h.AssignDriverToVehicle)
}
