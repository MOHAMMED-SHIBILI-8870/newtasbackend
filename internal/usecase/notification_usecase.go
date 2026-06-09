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
	return &NotificationUsecase{
		repo: r,
	}
}

func (u *NotificationUsecase) CreateNotification(
	ctx context.Context,
	notification *entity.Notification,
) error {

	if notification == nil {
		return errors.New("notification is required")
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

// ================= USER BOOKING NOTIFICATION =================

func (u *NotificationUsecase) CreateBookingNotification(
	ctx context.Context,
	userID uint,
	bookingID uint,
	message string,
) error {

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

// ================= ADMIN BOOKING NOTIFICATION =================

func (u *NotificationUsecase) CreateAdminBookingNotification(
	ctx context.Context,
	userID uint,
	bookingID uint,
	message string,
) error {

	return u.CreateNotification(ctx, &entity.Notification{
		UserID:    userID,
		Type:      "admin_booking",
		Title:     "New Booking",
		Message:   message,
		BookingID: &bookingID,
		IsRead:    false,
		IsAdmin:   true,
	})
}

// ================= AI REVIEW NOTIFICATION =================

func (u *NotificationUsecase) CreateAIReviewNotification(
	ctx context.Context,
	userID uint,
	aiRequestID uint,
	title string,
	message string,
) error {

	if userID == 0 || aiRequestID == 0 {
		return errors.New("invalid notification target")
	}

	return u.CreateNotification(ctx, &entity.Notification{
		UserID:          userID,
		Type:            "ai_request",
		Title:           title,
		Message:         message,
		AITripRequestID: &aiRequestID,
		IsRead:          false,
	})
}

// ================= USER GET =================

func (u *NotificationUsecase) GetUserNotifications(
	ctx context.Context,
	userID uint,
) ([]entity.Notification, error) {

	return u.repo.GetUserNotifications(ctx, userID)
}

// ================= ADMIN GET =================

func (u *NotificationUsecase) GetAdminNotifications(
	ctx context.Context,
) ([]entity.Notification, error) {

	return u.repo.GetAdminNotifications(ctx)
}

// ================= MARK AS READ =================

func (u *NotificationUsecase) MarkAsRead(
	ctx context.Context,
	role string,
	userID uint,
	notificationID uint,
) error {

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
