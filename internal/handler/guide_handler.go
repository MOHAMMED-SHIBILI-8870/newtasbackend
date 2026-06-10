package handler

import (
	"backend/internal/dto"
	"backend/internal/middleware"
	"backend/internal/response"
	"backend/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type GuideHandler struct {
	usecase *usecase.GuideUsecase
}

func NewGuideHandler(u *usecase.GuideUsecase) *GuideHandler {
	return &GuideHandler{
		usecase: u,
	}
}

func (h *GuideHandler) GetProfile(c *fiber.Ctx) error {
	userID := middleware.GetAuthUserID(c)

	guide, err := h.usecase.GetProfile(
		c.Context(),
		userID,
	)
	if err != nil {
		return response.Fail(
			c,
			fiber.StatusNotFound,
			"guide not found",
			err,
		)
	}

	return response.Success(
		c,
		fiber.StatusOK,
		"profile fetched",
		guide,
	)
}

func (h *GuideHandler) UpdateProfile(c *fiber.Ctx) error {
	var req dto.UpdateGuideProfileInput

	if err := c.BodyParser(&req); err != nil {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			"invalid request body",
			err,
		)
	}

	userID := middleware.GetAuthUserID(c)

	err := h.usecase.UpdateProfile(
		c.Context(),
		userID,
		req,
	)
	if err != nil {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			"failed to update profile",
			err,
		)
	}

	return response.Success(
		c,
		fiber.StatusOK,
		"profile updated successfully",
		nil,
	)
}