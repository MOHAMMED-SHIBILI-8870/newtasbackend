package routes

import (
	"backend/internal/handler"

	"github.com/gofiber/fiber/v2"
)

func DriverRoutes(
	app fiber.Router,
	h *handler.DriverHandler,
	auth fiber.Handler,
	permission func(...string) fiber.Handler,
) {
	// =====================================
	// ADMIN ENDPOINTS
	// =====================================
	adminDrivers := app.Group("/admin/drivers", auth)
	adminDrivers.Post("/", permission("manage_users"), h.CreateDriver)
	adminDrivers.Get("/", permission("manage_users"), h.ListDrivers)
	adminDrivers.Get("/:id", permission("manage_users"), h.GetDriverByID)
	adminDrivers.Put("/:id", permission("manage_users"), h.UpdateDriver)
	adminDrivers.Delete("/:id", permission("manage_users"), h.DeleteDriver)
	adminDrivers.Put("/:id/assign-vehicle", permission("manage_users"), h.AssignVehicle)

	// Admin Booking Assignment
	adminBookings := app.Group("/admin/bookings", auth)
	adminBookings.Put("/:id/assign-driver", permission("manage_bookings"), h.AssignDriverToBooking)

	// =====================================
	// DRIVER PORTAL ENDPOINTS
	// =====================================
	driverPortal := app.Group("/driver", auth)
	driverPortal.Get("/dashboard", permission("view_driver_dashboard"), h.GetDriverDashboard)
	driverPortal.Get("/trips", permission("view_assigned_trips"), h.GetDriverTrips)
	driverPortal.Get("/trips/:id", permission("view_assigned_trips"), h.GetDriverTripByID)
	driverPortal.Patch("/trips/:id/status", permission("update_trip_status"), h.UpdateTripStatus)
	driverPortal.Get("/vehicle", permission("view_vehicle"), h.GetDriverVehicle)
	driverPortal.Post("/tracking/update", permission("access_tracking"), h.UpdateDriverTracking)
	driverPortal.Get("/schedule", permission("view_assigned_trips"), h.GetDriverTrips)
	driverPortal.Get("/profile", permission("manage_profile"), h.GetDriverProfile)
	driverPortal.Put("/profile", permission("manage_profile"), h.UpdateDriverProfile)
}
