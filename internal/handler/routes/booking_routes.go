package routes

import (
	"backend/internal/handler"
	"backend/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// BookingRoutes configures standalone routing handles for customer order cycles
func BookingRoutes(app fiber.Router, h *handler.BookingHandler) {

	// =====================================
	// PROTECTED USER BOOKING ACTIONS
	// =====================================
	// Requires standard account login validation
	userBookings := app.Group("/bookings", middleware.AuthMiddleware())

	userBookings.Post("/trip/:id", h.BookTrip)                  // Create a fresh booking from an admin template ID
	userBookings.Get("/my-orders", h.GetUserBookings)          // Retrieve list of packages purchased by the active user
	userBookings.Patch("/custom-plan/:id", h.UpdateUserBookingPlans) // Allow user to modify their specific itinerary copy

	// =====================================
	// ADMIN ORDER DASHBOARD MANAGEMENT
	// =====================================
	// Requires elevated administrative operational rights
	adminOrders := app.Group("/admin/orders", middleware.AuthMiddleware(), middleware.RoleMiddleware("admin"))

	adminOrders.Get("/all", h.GetAllOrders) // Master dashboard list tracking system-wide consumer sales
}