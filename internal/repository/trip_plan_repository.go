package repository

import (
	"backend/internal/entity"
	"context"
	"gorm.io/gorm"
)

type TripPlanRepository interface {
	Create(ctx context.Context, plan *entity.TripPlan) error
	GetByTripID(ctx context.Context, tripID uint) ([]entity.TripPlan, error)
	Delete(ctx context.Context, id uint) error
}

type tripPlanRepository struct {
	db *gorm.DB
}

func NewTripPlanRepository(db *gorm.DB) TripPlanRepository {
	return &tripPlanRepository{db: db}
}

func (r *tripPlanRepository) Create(ctx context.Context, plan *entity.TripPlan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *tripPlanRepository) GetByTripID(ctx context.Context, tripID uint) ([]entity.TripPlan, error) {

	var plans []entity.TripPlan

	err := r.db.WithContext(ctx).
		Where("trip_id = ?", tripID).
		Order("day_number ASC").
		Order("start_time ASC").
		Find(&plans).Error

	return plans, err
}

func (r *tripPlanRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.TripPlan{}, id).Error
}
