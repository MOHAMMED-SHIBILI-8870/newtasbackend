package handler

import (
	"backend/internal/middleware"
	"backend/internal/response"
	"backend/internal/usecase"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type SupportHandler struct {
	usecase *usecase.SupportUsecase
}

func NewSupportHandler(u *usecase.SupportUsecase) *SupportHandler {
	return &SupportHandler{usecase: u}
}

type CreateSupportReq struct {
	Subject     string `json:"subject"`
	Description string `json:"description"`
}

func (h *SupportHandler) CreateRequest(c *fiber.Ctx) error {
	var input CreateSupportReq
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	userID := middleware.GetAuthUserID(c)
	req, err := h.usecase.CreateRequest(c.Context(), userID, input.Subject, input.Description)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "failed to create support request", err)
	}

	return response.Success(c, fiber.StatusCreated, "support request created", req)
}

func (h *SupportHandler) ListMyRequests(c *fiber.Ctx) error {
	userID := middleware.GetAuthUserID(c)
	reqs, err := h.usecase.ListMyRequests(c.Context(), userID)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to load requests", err)
	}
	return response.Success(c, fiber.StatusOK, "requests loaded", reqs)
}

func (h *SupportHandler) ListAllRequests(c *fiber.Ctx) error {
	role := middleware.GetAuthRole(c)
	status := c.Query("status")
	reqs, err := h.usecase.ListAllRequests(c.Context(), status, role)
	if err != nil {
		return response.Fail(c, fiber.StatusForbidden, "access denied", err)
	}
	return response.Success(c, fiber.StatusOK, "requests loaded", reqs)
}

type AssignAgentReq struct {
	AgentID uint `json:"agent_id"`
}

func (h *SupportHandler) AssignAgent(c *fiber.Ctx) error {
	reqID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid id", err)
	}

	var input AssignAgentReq
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid body", err)
	}

	role := middleware.GetAuthRole(c)
	if err := h.usecase.AssignAgent(c.Context(), uint(reqID), input.AgentID, role); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "failed to assign agent", err)
	}

	return response.Success(c, fiber.StatusOK, "agent assigned", nil)
}

type UpdateStatusReq struct {
	Status string `json:"status"`
}

func (h *SupportHandler) UpdateStatus(c *fiber.Ctx) error {
	reqID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid id", err)
	}

	var input UpdateStatusReq
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid body", err)
	}

	role := middleware.GetAuthRole(c)
	if err := h.usecase.UpdateStatus(c.Context(), uint(reqID), input.Status, role); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "failed to update status", err)
	}

	return response.Success(c, fiber.StatusOK, "status updated", nil)
}
