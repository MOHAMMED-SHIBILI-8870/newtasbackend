package handler

import (
	"backend/internal/entity"
	"backend/internal/middleware"
	"backend/internal/response"
	"backend/internal/usecase"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type BookingHandler struct {
	usecase *usecase.BookingUsecase
}

func NewBookingHandler(u *usecase.BookingUsecase) *BookingHandler {
	return &BookingHandler{usecase: u}
}

func bookingErrorStatus(err error) int {
	if err == nil {
		return fiber.StatusBadRequest
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "not found"):
		return fiber.StatusNotFound
	case strings.Contains(msg, "access denied"):
		return fiber.StatusForbidden
	default:
		return fiber.StatusBadRequest
	}
}

func (h *BookingHandler) BookTrip(c *fiber.Ctx) error {
	tripID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid trip id", err)
	}

	var input struct {
		Seats      int    `json:"seats"`
		CouponCode string `json:"coupon_code"`
	}
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&input); err != nil {
			return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
		}
	}

	userID := middleware.GetAuthUserID(c)
	if userID == 0 {
		return response.Fail(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	booking, err := h.usecase.BookTrip(c.Context(), uint(tripID), userID, input.Seats, input.CouponCode)
	if err != nil {
		return response.Fail(c, bookingErrorStatus(err), "failed to create booking", err)
	}

	return response.Success(c, fiber.StatusCreated, "booking created successfully", booking)
}

func (h *BookingHandler) GetUserBookings(c *fiber.Ctx) error {
	userID := middleware.GetAuthUserID(c)
	if userID == 0 {
		return response.Fail(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	bookings, err := h.usecase.GetUserBookings(c.Context(), userID)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to load bookings", err)
	}

	return response.Success(c, fiber.StatusOK, "bookings loaded successfully", bookings)
}

func (h *BookingHandler) UpdateUserBookingPlans(c *fiber.Ctx) error {
	bookingID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid booking id", err)
	}

	var input entity.UpdateBookingPlanInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	userID := middleware.GetAuthUserID(c)
	if userID == 0 {
		return response.Fail(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	if err := h.usecase.UpdateUserBookingPlans(c.Context(), uint(bookingID), userID, input.Plans); err != nil {
		return response.Fail(c, bookingErrorStatus(err), "failed to update booking plans", err)
	}

	return response.Success(c, fiber.StatusOK, "booking plans updated successfully", nil)
}

func (h *BookingHandler) GetAllOrders(c *fiber.Ctx) error {
	orders, err := h.usecase.GetAllOrders(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to load orders", err)
	}

	return response.Success(c, fiber.StatusOK, "orders loaded successfully", orders)
}
