package handler

import (
	"backend/internal/response"
	"backend/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type AnalyticsHandler struct {
	usecase *usecase.AnalyticsUsecase
}

func NewAnalyticsHandler(usecase *usecase.AnalyticsUsecase) *AnalyticsHandler {
	return &AnalyticsHandler{usecase: usecase}
}

func (h *AnalyticsHandler) GetDashboard(c *fiber.Ctx) error {
	role := c.Locals("role")
	roleStr := ""
	if role != nil {
		roleStr = role.(string)
	}

	metrics, err := h.usecase.GetDashboardMetrics(c.Context(), roleStr)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to load analytics", err)
	}

	return response.Success(c, fiber.StatusOK, "analytics loaded successfully", metrics)
}
