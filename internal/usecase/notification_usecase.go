package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"errors"
)

type NotificationUsecase struct {
	repo repository.NotificationRepository
}

func NewNotificationUsecase(r repository.NotificationRepository) *NotificationUsecase {
	return &NotificationUsecase{repo: r}
}

func (u *NotificationUsecase) CreateNotification(ctx context.Context, notification *entity.Notification) error {
	if notification == nil {
		return errors.New("notification is required")
	}
	if notification.UserID == 0 {
		return errors.New("invalid notification target")
	}
	if notification.Title == "" {
		notification.Title = "Notification"
	}
	if notification.Type == "" {
		notification.Type = "general"
	}
	if notification.Message == "" {
		return errors.New("notification message is required")
	}

	return u.repo.Create(ctx, notification)
}

func (u *NotificationUsecase) CreateBookingNotification(ctx context.Context, userID, bookingID uint, message string) error {
	if userID == 0 || bookingID == 0 {
		return errors.New("invalid notification target")
	}

	return u.CreateNotification(ctx, &entity.Notification{
		UserID:    userID,
		Type:      "booking_created",
		Title:     "Booking confirmed",
		Message:   message,
		BookingID: &bookingID,
		IsRead:    false,
	})
}

func (u *NotificationUsecase) CreateAIReviewNotification(ctx context.Context, userID uint, aiRequestID uint, title string, message string) error {
	if userID == 0 || aiRequestID == 0 {
		return errors.New("invalid notification target")
	}

	return u.CreateNotification(ctx, &entity.Notification{
		UserID:          userID,
		Type:            "ai_request",
		Title:           title,
		Message:         message,
		AITripRequestID:  &aiRequestID,
		IsRead:          false,
	})
}

func (u *NotificationUsecase) GetNotifications(ctx context.Context, role string, userID uint) ([]entity.Notification, error) {
	if role == "admin" {
		return u.repo.GetAll(ctx)
	}
	return u.repo.GetByUserID(ctx, userID)
}

func (u *NotificationUsecase) MarkAsRead(ctx context.Context, role string, userID, notificationID uint) error {
	notification, err := u.repo.GetByID(ctx, notificationID)
	if err != nil {
		return err
	}

	if role != "admin" && notification.UserID != userID {
		return errors.New("access denied")
	}

	if notification.IsRead {
		return nil
	}

	return u.repo.MarkAsRead(ctx, notificationID)
}
