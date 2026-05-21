package handler

import (
	"backend/internal/entity"
	"backend/internal/usecase"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type BookingHandler struct {
	usecase *usecase.BookingUsecase
}

func NewBookingHandler(u *usecase.BookingUsecase) *BookingHandler {
	return &BookingHandler{usecase: u}
}

// Internal package helper extraction targeting Fiber middleware parameters
func fetchContextUserID(c *fiber.Ctx) uint {
	if v, ok := c.Locals("user_id").(uint); ok { return v }
	if v, ok := c.Locals("user_id").(float64); ok { return uint(v) }
	return 0
}

func (h *BookingHandler) BookTrip(c *fiber.Ctx) error {
	tripID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid trip parameter key reference ID"})
	}

	userID := fetchContextUserID(c)
	if userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unverified authentication lifecycle identity context"})
	}

	booking, err := h.usecase.BookTrip(c.Context(), uint(tripID), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	
	return c.Status(fiber.StatusCreated).JSON(booking)
}

func (h *BookingHandler) GetUserBookings(c *fiber.Ctx) error {
	userID := fetchContextUserID(c)
	bookings, err := h.usecase.GetUserBookings(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(bookings)
}

func (h *BookingHandler) UpdateUserBookingPlans(c *fiber.Ctx) error {
	bookingID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Malformed structural routing identifier sequence"})
	}

	var input entity.UpdateBookingPlanInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Could not map invalid request array fields payload"})
	}
	
	userID := fetchContextUserID(c)
	err = h.usecase.UpdateUserBookingPlans(c.Context(), uint(bookingID), userID, input.Plans)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}
	
	return c.JSON(fiber.Map{"message": "Personal localized itinerary configuration updated smoothly"})
}

func (h *BookingHandler) GetAllOrders(c *fiber.Ctx) error {
	orders, err := h.usecase.GetAllOrders(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(200).JSON(orders)
}