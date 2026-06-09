package handler

import (
	"backend/internal/entity"
	"backend/internal/response"
	"backend/internal/usecase"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type TripSlotHandler struct {
	usecase *usecase.TripSlotUsecase
}

func NewTripSlotHandler(u *usecase.TripSlotUsecase) *TripSlotHandler {
	return &TripSlotHandler{usecase: u}
}

func tripSlotStatusFromErr(err error) int {
	if err == nil {
		return fiber.StatusBadRequest
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "not found"):
		return fiber.StatusNotFound
	case strings.Contains(msg, "already assigned"), strings.Contains(msg, "already booked"), strings.Contains(msg, "duplicate"), strings.Contains(msg, "overlap"), strings.Contains(msg, "full"), strings.Contains(msg, "not available"):
		return fiber.StatusConflict
	case strings.Contains(msg, "access denied"):
		return fiber.StatusForbidden
	case strings.Contains(msg, "required"), strings.Contains(msg, "invalid"), strings.Contains(msg, "cannot"):
		return fiber.StatusBadRequest
	default:
		return fiber.StatusUnprocessableEntity
	}
}

func (h *TripSlotHandler) CreateSlot(c *fiber.Ctx) error {
	var slot entity.TripSlot
	if err := c.BodyParser(&slot); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	created, err := h.usecase.CreateSlot(c.Context(), &slot)
	if err != nil {
		return response.Fail(c, tripSlotStatusFromErr(err), "failed to create slot", err)
	}

	return response.Success(c, fiber.StatusCreated, "slot created successfully", created)
}

func (h *TripSlotHandler) ListSlots(c *fiber.Ctx) error {
	slots, err := h.usecase.ListSlots(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to load slots", err)
	}

	return response.Success(c, fiber.StatusOK, "slots loaded successfully", slots)
}

func (h *TripSlotHandler) GetSlotByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid slot id", err)
	}

	slot, err := h.usecase.GetSlotByID(c.Context(), uint(id))
	if err != nil {
		return response.Fail(c, tripSlotStatusFromErr(err), "failed to load slot", err)
	}

	return response.Success(c, fiber.StatusOK, "slot loaded successfully", slot)
}

func (h *TripSlotHandler) UpdateSlot(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid slot id", err)
	}

	var input entity.UpdateTripSlotInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	slot, err := h.usecase.UpdateSlot(c.Context(), uint(id), input)
	if err != nil {
		return response.Fail(c, tripSlotStatusFromErr(err), "failed to update slot", err)
	}

	return response.Success(c, fiber.StatusOK, "slot updated successfully", slot)
}

func (h *TripSlotHandler) DeleteSlot(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid slot id", err)
	}

	if err := h.usecase.DeleteSlot(c.Context(), uint(id)); err != nil {
		return response.Fail(c, tripSlotStatusFromErr(err), "failed to delete slot", err)
	}

	return response.Success(c, fiber.StatusOK, "slot deleted successfully", nil)
}

func (h *TripSlotHandler) GetUpcomingSlotsByTripID(c *fiber.Ctx) error {
	tripID, err := strconv.Atoi(c.Params("trip_id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid trip id", err)
	}

	slots, err := h.usecase.GetUpcomingSlotsByTripID(c.Context(), uint(tripID))
	if err != nil {
		return response.Fail(c, tripSlotStatusFromErr(err), "failed to load trip slots", err)
	}

	return response.Success(c, fiber.StatusOK, "trip slots loaded successfully", slots)
}
