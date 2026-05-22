package handler

import (
	"backend/internal/dto"
	"backend/internal/middleware"
	"backend/internal/response"
	"backend/internal/usecase"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type NotificationHandler struct {
	usecase *usecase.NotificationUsecase
}

func NewNotificationHandler(u *usecase.NotificationUsecase) *NotificationHandler {
	return &NotificationHandler{usecase: u}
}

func notificationStatusFromErr(err error) int {
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

func (h *NotificationHandler) GetNotifications(c *fiber.Ctx) error {
	userID := middleware.GetAuthUserID(c)
	role := middleware.GetAuthRole(c)

	notifications, err := h.usecase.GetNotifications(c.Context(), role, userID)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to load notifications", err)
	}

	items := make([]dto.NotificationResponse, 0, len(notifications))
	for _, n := range notifications {
		items = append(items, dto.NotificationResponse{
			ID:              n.ID,
			Type:            n.Type,
			Title:           n.Title,
			Message:         n.Message,
			UserID:          n.UserID,
			BookingID:       n.BookingID,
			AITripRequestID: n.AITripRequestID,
			IsRead:          n.IsRead,
			CreatedAt:       n.CreatedAt,
		})
	}

	return response.Success(c, fiber.StatusOK, "notifications loaded successfully", items)
}

func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid notification id", err)
	}

	userID := middleware.GetAuthUserID(c)
	role := middleware.GetAuthRole(c)

	if err := h.usecase.MarkAsRead(c.Context(), role, userID, uint(id)); err != nil {
		return response.Fail(c, notificationStatusFromErr(err), "unable to mark notification as read", err)
	}

	return response.Success(c, fiber.StatusOK, "notification marked as read", nil)
}
