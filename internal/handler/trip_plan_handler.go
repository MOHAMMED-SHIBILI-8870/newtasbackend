package handler

import (
	"backend/internal/entity"
	"backend/internal/response"
	"backend/internal/usecase"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type TripPlanHandler struct {
	usecase *usecase.TripPlanUsecase
}

func NewTripPlanHandler(u *usecase.TripPlanUsecase) *TripPlanHandler {
	return &TripPlanHandler{usecase: u}
}

func tripPlanStatusFromErr(err error) int {
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

func (h *TripPlanHandler) CreateTripPlan(c *fiber.Ctx) error {

	var plans []entity.TripPlan

	if err := c.BodyParser(&plans); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	createdPlans, err := h.usecase.CreateTripPlans(c.Context(), plans)
	if err != nil {
		return response.Fail(
			c,
			tripPlanStatusFromErr(err),
			"failed to create trip plans",
			err,
		)
	}

	return response.Success(
		c,
		fiber.StatusCreated,
		"trip plans created successfully",
		createdPlans,
	)
}

func (h *TripPlanHandler) GetTripPlanByID(c *fiber.Ctx) error {

	id, err := strconv.Atoi(c.Params("trip_id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid trip id", err)
	}

	plan, err := h.usecase.GetTripPlans(c.Context(), uint(id))
	if err != nil {
		return response.Fail(c, fiber.StatusNotFound, "trip plans not found", err)
	}

	return response.Success(c, fiber.StatusOK, "trip plans loaded successfully", plan)
}

func (h *TripPlanHandler) DeleteTripPlan(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid trip plan id", err)
	}

	if err := h.usecase.DeleteTripPlan(c.Context(), uint(id)); err != nil {
		return response.Fail(c, tripPlanStatusFromErr(err), "failed to delete trip plan", err)
	}

	return response.Success(c, fiber.StatusOK, "trip plan deleted successfully", nil)
}
