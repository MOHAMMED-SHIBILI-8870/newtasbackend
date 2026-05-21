package handler

import (
	"backend/internal/entity"
	"backend/internal/usecase"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type TripHandler struct {
	usecase *usecase.TripUsecase
}

func NewTripHandler(u *usecase.TripUsecase) *TripHandler {
	return &TripHandler{usecase: u}
}

func (h *TripHandler) CreateTrip(c *fiber.Ctx) error {
	var trip entity.Trip

	if err := c.BodyParser(&trip); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Malformed JSON layout format input",
		})
	}

	if err := h.usecase.CreateTrip(c.Context(), &trip); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(trip)
}

func (h *TripHandler) GetTripByName(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid trip parameter query",
		})
	}

	trip, err := h.usecase.GetTripByName(c.Context(), name)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Targeted package not found",
		})
	}

	return c.JSON(trip)
}

func (h *TripHandler) GetAllTrips(c *fiber.Ctx) error {
	trips, err := h.usecase.GetAllTrips(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to pull trips collection from store",
		})
	}

	return c.Status(200).JSON(trips)
}

func (h *TripHandler) UpdateTrip(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid administrative key target ID",
		})
	}

	var input entity.UpdateTripInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid modification structural schema parser match",
		})
	}

	if err := h.usecase.UpdateTrip(c.Context(), uint(id), input); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Trip package data records updated successfully",
	})
}

func (h *TripHandler) DeleteTrip(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Target sequence index mismatch parsing",
		})
	}

	if err := h.usecase.DeleteTrip(c.Context(), uint(id)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "System failed executing resource extraction target",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Package safely deleted from application records",
	})
}