package repository

import (
	"backend/internal/entity"
	"context"

	"gorm.io/gorm"
)

type TrackingRepository interface {
	Create(ctx context.Context, tracking *entity.Tracking) error
	Update(ctx context.Context, tracking *entity.Tracking) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*entity.Tracking, error)
	GetLatestByBookingID(ctx context.Context, bookingID uint) (*entity.Tracking, error)
	GetByBookingID(ctx context.Context, bookingID uint) ([]entity.Tracking, error)
	GetAll(ctx context.Context) ([]entity.Tracking, error)
}

type trackingRepository struct {
	db *gorm.DB
}

func NewTrackingRepository(db *gorm.DB) TrackingRepository {
	return &trackingRepository{db: db}
}

func (r *trackingRepository) Create(ctx context.Context, tracking *entity.Tracking) error {
	return r.db.WithContext(ctx).Create(tracking).Error
}

func (r *trackingRepository) Update(ctx context.Context, tracking *entity.Tracking) error {
	return r.db.WithContext(ctx).Save(tracking).Error
}

func (r *trackingRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Tracking{}, id).Error
}

func (r *trackingRepository) GetByID(ctx context.Context, id uint) (*entity.Tracking, error) {
	var tracking entity.Tracking
	err := r.db.WithContext(ctx).First(&tracking, id).Error
	if err != nil {
		return nil, err
	}
	return &tracking, nil
}

func (r *trackingRepository) GetLatestByBookingID(ctx context.Context, bookingID uint) (*entity.Tracking, error) {
	var tracking entity.Tracking
	err := r.db.WithContext(ctx).
		Where("booking_id = ?", bookingID).
		Order("updated_at DESC, id DESC").
		First(&tracking).Error
	if err != nil {
		return nil, err
	}
	return &tracking, nil
}

func (r *trackingRepository) GetByBookingID(ctx context.Context, bookingID uint) ([]entity.Tracking, error) {
	var trackings []entity.Tracking
	err := r.db.WithContext(ctx).
		Where("booking_id = ?", bookingID).
		Order("updated_at DESC, id DESC").
		Find(&trackings).Error
	return trackings, err
}

func (r *trackingRepository) GetAll(ctx context.Context) ([]entity.Tracking, error) {
	var trackings []entity.Tracking
	err := r.db.WithContext(ctx).Order("updated_at DESC, id DESC").Find(&trackings).Error
	return trackings, err
}
