package handler

import (
	"backend/internal/entity"
	"backend/internal/usecase"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type TripPlanHandler struct {
	usecase *usecase.TripPlanUsecase
}

func NewTripPlanHandler(u *usecase.TripPlanUsecase) *TripPlanHandler {
	return &TripPlanHandler{usecase: u}
}

func (h *TripPlanHandler) CreateTripPlan(c *fiber.Ctx) error {
	var plan entity.TripPlan

	// Parse request
	if err := c.BodyParser(&plan); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Basic validation (NEW + IMPORTANT)
	if plan.TripID == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "trip_id is required",
		})
	}

	if plan.DayNumber <= 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "day_number must be greater than 0",
		})
	}

	if plan.Title == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "title is required",
		})
	}

	// NEW validation for cost (optional but recommended)
	if plan.Cost < 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "cost cannot be negative",
		})
	}

	// Call usecase
	if err := h.usecase.CreateTripPlan(c.Context(), &plan); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(201).JSON(plan)
}

// GET ITINERARY BY TRIP ID
func (h *TripPlanHandler) GetTripPlanByID(c *fiber.Ctx) error {

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	plan, err := h.usecase.GetTripPlans(c.Context(), uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "plan not found",
		})
	}

	return c.JSON(plan)
}

// DELETE PLAN
func (h *TripPlanHandler) DeleteTripPlan(c *fiber.Ctx) error {

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	if err := h.usecase.DeleteTripPlan(c.Context(), uint(id)); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(204)
}
