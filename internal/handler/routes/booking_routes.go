package routes

import (
	"backend/internal/handler"

	"github.com/gofiber/fiber/v2"
)

// BookingRoutes configures standalone routing handles for customer order cycles
func BookingRoutes(app fiber.Router, h *handler.BookingHandler, auth fiber.Handler, admin fiber.Handler) {

	// =====================================
	// PROTECTED USER BOOKING ACTIONS
	// =====================================
	// Requires standard account login validation
	userBookings := app.Group("/bookings", auth)

	userBookings.Get("/my-orders", h.GetUserBookings) 
	userBookings.Get("/:id", h.GetBookingByID)
	userBookings.Post("/slot/:slot_id", h.BookSlot)                  // Create a booking against a specific trip slot
	userBookings.Post("/trip/:id", h.BookTrip)               // Retrieve list of packages purchased by the active user
	userBookings.Patch("/custom-plan/:id", h.UpdateUserBookingPlans) // Allow user to modify their specific itinerary copy

	// =====================================
	// ADMIN ORDER DASHBOARD MANAGEMENT
	// =====================================
	// Requires elevated administrative operational rights
	adminOrders := app.Group("/admin/orders", auth, admin)

	adminOrders.Get("/all", h.GetAllOrders) // Master dashboard list tracking system-wide consumer sales
}
