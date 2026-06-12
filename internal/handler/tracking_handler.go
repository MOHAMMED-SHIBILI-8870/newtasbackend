package handler

import (
	"backend/internal/dto"
	"backend/internal/middleware"
	"backend/internal/response"
	"backend/internal/usecase"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type TrackingHandler struct {
	usecase *usecase.TrackingUsecase
}

func NewTrackingHandler(usecase *usecase.TrackingUsecase) *TrackingHandler {
	return &TrackingHandler{usecase: usecase}
}

func trackingStatusFromErr(err error) int {
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

func mapTrackingResponse(tracking dto.TrackingResponse) dto.TrackingResponse {
	return tracking
}

func (h *TrackingHandler) UpdateLocation(c *fiber.Ctx) error {
	var input dto.TrackingUpdateRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	tracking, err := h.usecase.UpdateLocation(c.Context(), input.BookingID, *input.VehicleID, input.Latitude, input.Longitude)
	if err != nil {
		return response.Fail(c, trackingStatusFromErr(err), "failed to update tracking", err)
	}

	return response.Success(c, fiber.StatusCreated, "tracking updated successfully", dto.TrackingResponse{
		ID:        tracking.ID,
		BookingID: tracking.BookingID,
		VehicleID: tracking.VehicleID,
		DriverID:  tracking.DriverID,
		Type:      tracking.Type,
		Latitude:  tracking.Latitude,
		Longitude: tracking.Longitude,
		CreatedAt: tracking.CreatedAt,
		UpdatedAt: tracking.UpdatedAt,
	})
}

func (h *TrackingHandler) GetLatestByBookingID(c *fiber.Ctx) error {
	bookingID, err := strconv.Atoi(c.Params("booking_id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid booking id", err)
	}

	userID := middleware.GetAuthUserID(c)
	role := middleware.GetAuthRole(c)
	tracking, err := h.usecase.GetLatestForUser(c.Context(), userID, uint(bookingID), role)
	if err != nil {
		return response.Fail(c, trackingStatusFromErr(err), "failed to load tracking", err)
	}
	if tracking == nil {
		return response.Fail(c, fiber.StatusNotFound, "tracking not found", nil)
	}

	return response.Success(c, fiber.StatusOK, "tracking loaded successfully", dto.TrackingResponse{
		ID:        tracking.ID,
		BookingID: tracking.BookingID,
		VehicleID: tracking.VehicleID,
		DriverID:  tracking.DriverID,
		Type:      tracking.Type,
		Latitude:  tracking.Latitude,
		Longitude: tracking.Longitude,
		CreatedAt: tracking.CreatedAt,
		UpdatedAt: tracking.UpdatedAt,
	})
}

func (h *TrackingHandler) GetTrackingsByBookingID(c *fiber.Ctx) error {
	bookingID, err := strconv.Atoi(c.Params("booking_id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid booking id", err)
	}

	userID := middleware.GetAuthUserID(c)
	role := middleware.GetAuthRole(c)
	trackings, err := h.usecase.GetTrackingsForUser(c.Context(), userID, uint(bookingID), role)
	if err != nil {
		return response.Fail(c, trackingStatusFromErr(err), "failed to load tracking history", err)
	}

	items := make([]dto.TrackingResponse, 0, len(trackings))
	for _, tracking := range trackings {
		items = append(items, dto.TrackingResponse{
			ID:        tracking.ID,
			BookingID: tracking.BookingID,
			VehicleID: tracking.VehicleID,
			DriverID:  tracking.DriverID,
			Type:      tracking.Type,
			Latitude:  tracking.Latitude,
			Longitude: tracking.Longitude,
			CreatedAt: tracking.CreatedAt,
			UpdatedAt: tracking.UpdatedAt,
		})
	}

	return response.Success(c, fiber.StatusOK, "tracking history loaded successfully", items)
}

func (h *TrackingHandler) GetAllTracking(c *fiber.Ctx) error {
	trackings, err := h.usecase.GetAll(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to load tracking", err)
	}

	items := make([]dto.TrackingResponse, 0, len(trackings))
	for _, tracking := range trackings {
		items = append(items, dto.TrackingResponse{
			ID:        tracking.ID,
			BookingID: tracking.BookingID,
			VehicleID: tracking.VehicleID,
			DriverID:  tracking.DriverID,
			Type:      tracking.Type,
			Latitude:  tracking.Latitude,
			Longitude: tracking.Longitude,
			CreatedAt: tracking.CreatedAt,
			UpdatedAt: tracking.UpdatedAt,
		})
	}

	return response.Success(c, fiber.StatusOK, "tracking loaded successfully", items)
}

func (h *TrackingHandler) GetDashboard(c *fiber.Ctx) error {
	trackings, err := h.usecase.GetDashboard(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to load dashboard tracking", err)
	}

	items := make([]dto.TrackingResponse, 0, len(trackings))
	for _, tracking := range trackings {
		items = append(items, dto.TrackingResponse{
			ID:        tracking.ID,
			BookingID: tracking.BookingID,
			VehicleID: tracking.VehicleID,
			DriverID:  tracking.DriverID,
			Type:      tracking.Type,
			Latitude:  tracking.Latitude,
			Longitude: tracking.Longitude,
			CreatedAt: tracking.CreatedAt,
			UpdatedAt: tracking.UpdatedAt,
		})
	}

	return response.Success(c, fiber.StatusOK, "tracking dashboard loaded successfully", items)
}
