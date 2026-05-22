package repository

import (
	"backend/internal/entity"
	"context"

	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification *entity.Notification) error
	GetAll(ctx context.Context) ([]entity.Notification, error)
	GetByUserID(ctx context.Context, userID uint) ([]entity.Notification, error)
	GetByID(ctx context.Context, id uint) (*entity.Notification, error)
	MarkAsRead(ctx context.Context, id uint) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, notification *entity.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *notificationRepository) GetAll(ctx context.Context) ([]entity.Notification, error) {
	var notifications []entity.Notification
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Booking").
		Preload("AITripRequest").
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) GetByUserID(ctx context.Context, userID uint) ([]entity.Notification, error) {
	var notifications []entity.Notification
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Booking").
		Preload("AITripRequest").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) GetByID(ctx context.Context, id uint) (*entity.Notification, error) {
	var notification entity.Notification
	if err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Booking").
		Preload("AITripRequest").
		First(&notification, id).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&entity.Notification{}).
		Where("id = ?", id).
		Update("is_read", true).Error
}
