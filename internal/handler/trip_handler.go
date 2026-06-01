package handler

import (
	"backend/internal/entity"
	"backend/internal/response"
	"backend/internal/usecase"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"net/url"
)

type TripHandler struct {
	usecase *usecase.TripUsecase
}

func NewTripHandler(u *usecase.TripUsecase) *TripHandler {
	return &TripHandler{usecase: u}
}

func tripStatusFromErr(err error) int {
	if err == nil {
		return fiber.StatusBadRequest
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "not found"):
		return fiber.StatusNotFound
	case strings.Contains(msg, "required"), strings.Contains(msg, "invalid"):
		return fiber.StatusBadRequest
	default:
		return fiber.StatusUnprocessableEntity
	}
}

func (h *TripHandler) CreateTrip(c *fiber.Ctx) error {
	var trip entity.Trip
	if err := c.BodyParser(&trip); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	if err := h.usecase.CreateTrip(c.Context(), &trip); err != nil {
		return response.Fail(c, tripStatusFromErr(err), "failed to create trip", err)
	}

	return response.Success(c, fiber.StatusCreated, "trip created successfully", trip)
}

func (h *TripHandler) GetTripByName(c *fiber.Ctx) error {

	name, err := url.QueryUnescape(c.Params("name"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid trip name", err)
	}

	if name == "" {
		return response.Fail(c, fiber.StatusBadRequest, "trip name is required", nil)
	}

	trip, err := h.usecase.GetTripByName(c.Context(), name)
	if err != nil {
		return response.Fail(c, fiber.StatusNotFound, "trip not found", err)
	}

	return response.Success(c, fiber.StatusOK, "trip loaded successfully", trip)
}

func (h *TripHandler) GetAllTrips(c *fiber.Ctx) error {
	trips, err := h.usecase.GetAllTrips(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to load trips", err)
	}

	return response.Success(c, fiber.StatusOK, "trips loaded successfully", trips)
}

func (h *TripHandler) UpdateTrip(c *fiber.Ctx) error {

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid trip id", err)
	}

	var input entity.UpdateTripInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	if err := h.usecase.UpdateTrip(c.Context(), uint(id), input); err != nil {
		return response.Fail(c, tripStatusFromErr(err), "failed to update trip", err)
	}

	return response.Success(c, fiber.StatusOK, "trip updated successfully", nil)
}

func (h *TripHandler) DeleteTrip(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid trip id", err)
	}

	if err := h.usecase.DeleteTrip(c.Context(), uint(id)); err != nil {
		return response.Fail(c, tripStatusFromErr(err), "failed to delete trip", err)
	}

	return response.Success(c, fiber.StatusOK, "trip deleted successfully", nil)
}
