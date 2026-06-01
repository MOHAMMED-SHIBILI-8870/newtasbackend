package repository

import (
	"backend/internal/entity"
	"context"

	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification *entity.Notification) error

	GetUserNotifications(ctx context.Context, userID uint) ([]entity.Notification, error)

	GetAdminNotifications(ctx context.Context) ([]entity.Notification, error)

	GetByID(ctx context.Context, id uint) (*entity.Notification, error)

	MarkAsRead(ctx context.Context, id uint) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{
		db: db,
	}
}

func (r *notificationRepository) Create(
	ctx context.Context,
	notification *entity.Notification,
) error {

	return r.db.WithContext(ctx).
		Create(notification).Error
}

// ================= USER NOTIFICATIONS =================

func (r *notificationRepository) GetUserNotifications(
	ctx context.Context,
	userID uint,
) ([]entity.Notification, error) {

	var notifications []entity.Notification

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&notifications).Error

	return notifications, err
}

// ================= ADMIN NOTIFICATIONS =================

func (r *notificationRepository) GetAdminNotifications(
	ctx context.Context,
) ([]entity.Notification, error) {

	var notifications []entity.Notification

	err := r.db.WithContext(ctx).
		Where("type = ?", "admin_booking").
		Order("created_at DESC").
		Find(&notifications).Error

	return notifications, err
}

func (r *notificationRepository) GetByID(
	ctx context.Context,
	id uint,
) (*entity.Notification, error) {

	var notification entity.Notification

	err := r.db.WithContext(ctx).
		First(&notification, id).Error

	if err != nil {
		return nil, err
	}

	return &notification, nil
}

func (r *notificationRepository) MarkAsRead(
	ctx context.Context,
	id uint,
) error {

	return r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("id = ?", id).
		Update("is_read", true).Error
}
