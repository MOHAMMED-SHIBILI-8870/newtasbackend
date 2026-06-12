package handler

import (
	"backend/internal/entity"
	"backend/internal/middleware"
	"backend/internal/response"
	"backend/internal/usecase"
	"strconv"
	"strings"
	"time"

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
	case strings.Contains(msg, "already booked"), strings.Contains(msg, "duplicate"), strings.Contains(msg, "full"), strings.Contains(msg, "insufficient"), strings.Contains(msg, "overlap"), strings.Contains(msg, "not available"):
		return fiber.StatusConflict
	case strings.Contains(msg, "access denied"):
		return fiber.StatusForbidden
	default:
		if strings.Contains(msg, "required") || strings.Contains(msg, "invalid") {
			return fiber.StatusBadRequest
		}
		return fiber.StatusInternalServerError
	}
}

type bookingInput struct {
	Seats       int    `json:"seats"`
	CouponCode  string `json:"coupon_code"`
	BookingType string `json:"booking_type"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
}

func (h *BookingHandler) BookTrip(c *fiber.Ctx) error {
	tripID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid trip id", err)
	}

	var input bookingInput
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&input); err != nil {
			return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
		}
	}

	userID := middleware.GetAuthUserID(c)
	if userID == 0 {
		return response.Fail(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	var startDate, endDate *time.Time
	if input.StartDate != "" {
		if t, err := time.Parse(time.RFC3339, input.StartDate); err == nil {
			startDate = &t
		} else if t, err := time.Parse("2006-01-02", input.StartDate); err == nil {
			startDate = &t
		}
	}
	if input.EndDate != "" {
		if t, err := time.Parse(time.RFC3339, input.EndDate); err == nil {
			endDate = &t
		} else if t, err := time.Parse("2006-01-02", input.EndDate); err == nil {
			endDate = &t
		}
	}

	booking, err := h.usecase.BookTrip(c.Context(), uint(tripID), userID, input.Seats, input.CouponCode, startDate, endDate)
	if err != nil {
		return response.Fail(c, bookingErrorStatus(err), "failed to create booking", err)
	}

	return response.Success(c, fiber.StatusCreated, "booking created successfully", booking)
}

func (h *BookingHandler) BookSlot(c *fiber.Ctx) error {
	slotID, err := strconv.Atoi(c.Params("slot_id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid slot id", err)
	}

	var input bookingInput
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&input); err != nil {
			return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
		}
	}

	userID := middleware.GetAuthUserID(c)
	if userID == 0 {
		return response.Fail(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	var startDate, endDate *time.Time
	if input.StartDate != "" {
		if t, err := time.Parse(time.RFC3339, input.StartDate); err == nil {
			startDate = &t
		} else if t, err := time.Parse("2006-01-02", input.StartDate); err == nil {
			startDate = &t
		}
	}
	if input.EndDate != "" {
		if t, err := time.Parse(time.RFC3339, input.EndDate); err == nil {
			endDate = &t
		} else if t, err := time.Parse("2006-01-02", input.EndDate); err == nil {
			endDate = &t
		}
	}

	booking, err := h.usecase.BookSlot(c.Context(), uint(slotID), userID, input.Seats, input.CouponCode, input.BookingType, startDate, endDate)
	if err != nil {
		return response.Fail(c, bookingErrorStatus(err), "failed to create booking", err)
	}

	return response.Success(c, fiber.StatusCreated, "booking created successfully", booking)
}

func (h *BookingHandler) GetBookingByID(c *fiber.Ctx) error {
	bookingID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(
			c, 
			fiber.StatusBadRequest, 
			"invalid booking id format", 
			err,
		)
	}

	userID := middleware.GetAuthUserID(c)
	role := middleware.GetAuthRole(c)
	if userID == 0 {
		return response.Fail(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	// Calls the clean usecase layer wrapper with ownership validation
	booking, err := h.usecase.GetBookingByID(c.Context(), uint(bookingID), userID, role)
	if err != nil {
		return response.Fail(
			c, 
			bookingErrorStatus(err), 
			"failed to fetch booking info", 
			err,
		)
	}

	return response.Success(
		c, 
		fiber.StatusOK, 
		"booking details retrieved successfully", 
		booking,
	)
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
