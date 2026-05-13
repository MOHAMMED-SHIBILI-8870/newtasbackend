package handler

import (
	"backend/internal/entity"
	"backend/internal/usecase"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type TripHandler struct {
	usecase *usecase.TripUsecase
}

func NewTripHandler(u *usecase.TripUsecase) *TripHandler {
	return &TripHandler{usecase: u}
}

// CreateTrip: POST /trips
func (h *TripHandler) CreateTrip(c *fiber.Ctx) error {
	var trip entity.Trip

	// 1. Parse JSON body from React
	if err := c.BodyParser(&trip); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON format",
		})
	}

	userVal := c.Locals("user_id")

	var userID uint
	switch v := userVal.(type) {
	case float64:
		userID = uint(v)
	case int:
		userID = uint(v)
	case uint:
		userID = v
	default:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid user_id type",
		})
	}

	trip.UserId = userID

	if trip.Duration <= 0 {
		return errors.New("duration must be at least 1 day")
	}

	// 3. Delegate to Usecase
	if err := h.usecase.CreateTrip(c.Context(), &trip); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(trip)
}

// GetTripByID: GET /trips/:id
func (h *TripHandler) GetTripByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	trip, err := h.usecase.GetTripDetails(c.Context(), uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Trip not found",
		})
	}

	return c.JSON(trip)
}

// GetUserTrips: GET /trips/my-trips
func (h *TripHandler) GetUserTrips(c *fiber.Ctx) error {
	userVal := c.Locals("user_id")

	var userID uint
	switch v := userVal.(type) {
	case float64:
		userID = uint(v)
	case int:
		userID = uint(v)
	case uint:
		userID = v
	default:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid user_id type",
		})
	}

	trips, err := h.usecase.GetTripsByOwner(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch trips",
		})
	}

	return c.JSON(trips)
}

// UpdateTrip: PATCH /trips/:id
// UpdateTrip: PATCH /admin/trips/:id
func (h *TripHandler) UpdateTrip(c *fiber.Ctx) error {

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid trip id",
		})
	}

	var input entity.UpdateTripInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid input",
		})
	}

	userVal := c.Locals("user_id")

	var userID uint

	switch v := userVal.(type) {
	case float64:
		userID = uint(v)

	case int:
		userID = uint(v)

	case uint:
		userID = v

	default:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid user_id type",
		})
	}

	if err := h.usecase.UpdateTrip(c.Context(),uint(id),input,userID); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "trip updated successfully",
	})
}

// DeleteTrip: DELETE /trips/:id
func (h *TripHandler) DeleteTrip(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid trip id",
		})
	}
	userVal := c.Locals("user_id")

	var userID uint
	switch v := userVal.(type) {
	case float64:
		userID = uint(v)
	case int:
		userID = uint(v)
	case uint:
		userID = v
	default:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid user_id type",
		})
	}

	if err := h.usecase.DeleteTrip(c.Context(), uint(id), userID); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Unauthorized or delete failed",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *TripHandler) GetAllTrips(c *fiber.Ctx) error {

	trips, err := h.usecase.GetAllTrips(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch trips",
		})
	}

	return c.JSON(trips)
}
