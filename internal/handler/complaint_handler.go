package handler

import (
	"backend/internal/dto"
	"backend/internal/entity"
	"backend/internal/middleware"
	"backend/internal/response"
	"backend/internal/usecase"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ComplaintHandler struct {
	usecase *usecase.ComplaintUsecase
}

func NewComplaintHandler(usecase *usecase.ComplaintUsecase) *ComplaintHandler {
	return &ComplaintHandler{usecase: usecase}
}

func complaintStatusFromErr(err error) int {
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

func mapComplaintResponse(complaint entity.Complaint) dto.ComplaintResponse {
	return dto.ComplaintResponse{
		ID:          complaint.ID,
		UserID:      complaint.UserID,
		BookingID:   complaint.BookingID,
		Title:       complaint.Title,
		Description: complaint.Description,
		Status:      complaint.Status,
		CreatedAt:   complaint.CreatedAt,
		UpdatedAt:   complaint.UpdatedAt,
	}
}

func (h *ComplaintHandler) CreateComplaint(c *fiber.Ctx) error {
	var input dto.ComplaintRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	userID := middleware.GetAuthUserID(c)
	complaint, err := h.usecase.CreateComplaint(c.Context(), userID, input.BookingID, input.Title, input.Description)
	if err != nil {
		return response.Fail(c, complaintStatusFromErr(err), "failed to create complaint", err)
	}

	return response.Success(c, fiber.StatusCreated, "complaint created successfully", mapComplaintResponse(*complaint))
}

func (h *ComplaintHandler) ListMyComplaints(c *fiber.Ctx) error {
	userID := middleware.GetAuthUserID(c)
	complaints, err := h.usecase.ListUserComplaints(c.Context(), userID)
	if err != nil {
		return response.Fail(c, complaintStatusFromErr(err), "failed to load complaints", err)
	}

	items := make([]dto.ComplaintResponse, 0, len(complaints))
	for _, complaint := range complaints {
		items = append(items, mapComplaintResponse(complaint))
	}

	return response.Success(c, fiber.StatusOK, "complaints loaded successfully", items)
}

func (h *ComplaintHandler) ListAllComplaints(c *fiber.Ctx) error {
	complaints, err := h.usecase.ListAllComplaints(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to load complaints", err)
	}

	items := make([]dto.ComplaintResponse, 0, len(complaints))
	for _, complaint := range complaints {
		items = append(items, mapComplaintResponse(complaint))
	}

	return response.Success(c, fiber.StatusOK, "complaints loaded successfully", items)
}

func (h *ComplaintHandler) GetComplaintByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid complaint id", err)
	}

	userID := middleware.GetAuthUserID(c)
	role := middleware.GetAuthRole(c)
	complaint, err := h.usecase.GetComplaintByIDForUser(c.Context(), userID, role, uint(id))
	if err != nil {
		return response.Fail(c, complaintStatusFromErr(err), "failed to load complaint", err)
	}
	if complaint == nil {
		return response.Fail(c, fiber.StatusNotFound, "complaint not found", nil)
	}

	return response.Success(c, fiber.StatusOK, "complaint loaded successfully", mapComplaintResponse(*complaint))
}

func (h *ComplaintHandler) UpdateComplaintStatus(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid complaint id", err)
	}

	var input dto.ComplaintStatusRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	if err := h.usecase.UpdateComplaintStatus(c.Context(), uint(id), input.Status); err != nil {
		return response.Fail(c, complaintStatusFromErr(err), "failed to update complaint status", err)
	}

	return response.Success(c, fiber.StatusOK, "complaint status updated successfully", nil)
}

func (h *ComplaintHandler) DeleteComplaint(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid complaint id", err)
	}

	if err := h.usecase.DeleteComplaint(c.Context(), uint(id)); err != nil {
		return response.Fail(c, complaintStatusFromErr(err), "failed to delete complaint", err)
	}

	return response.Success(c, fiber.StatusOK, "complaint deleted successfully", nil)
}
