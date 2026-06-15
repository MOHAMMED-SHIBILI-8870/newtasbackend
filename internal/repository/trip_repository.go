package repository

import (
	"backend/internal/entity"
	"context"

	"gorm.io/gorm"
)

type TripRepository interface {
	Create(ctx context.Context, trip *entity.Trip) error
	GetByID(ctx context.Context, id uint) (*entity.Trip, error)
	GetByName(ctx context.Context, name string) (*entity.Trip, error)
	GetAll(ctx context.Context) ([]entity.Trip, error)
	Update(ctx context.Context, trip *entity.Trip) error
	Delete(ctx context.Context, id uint) error
	SearchSimilarTrips(ctx context.Context,destination string) ([]entity.Trip, error)
}

type tripRepository struct {
	db *gorm.DB
}

func NewTripRepository(db *gorm.DB) TripRepository {
	return &tripRepository{db: db}
}

func (r *tripRepository) Create(ctx context.Context, trip *entity.Trip) error {
	// GORM automatically saves associated Plans if populated inside the struct
	return r.db.WithContext(ctx).Create(trip).Error
}

func (r *tripRepository) GetByID(ctx context.Context, id uint) (*entity.Trip, error) {
	var trip entity.Trip
	err := r.db.WithContext(ctx).
		Preload("Plans").
		Preload("PricingTiers").
		First(&trip, id).Error
	return &trip, err
}

func (r *tripRepository) GetAll(ctx context.Context) ([]entity.Trip, error) {
	var trips []entity.Trip
	err := r.db.WithContext(ctx).
		Preload("Plans").
		Preload("PricingTiers").
		Find(&trips).Error
	return trips, err
}

func (r *tripRepository) GetByName(ctx context.Context, name string) (*entity.Trip, error) {
	var trip entity.Trip
	err := r.db.WithContext(ctx).
		Preload("Plans").
		Preload("PricingTiers").
		Where("LOWER(\"from\") LIKE LOWER(?) OR LOWER(\"to\") LIKE LOWER(?)", "%"+name+"%", "%"+name+"%").
		First(&trip).Error

	if err != nil {
		return nil, err
	}
	return &trip, nil
}

func (r *tripRepository) Update(ctx context.Context, trip *entity.Trip) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Drop existing plan associations first to avoid duplicate appending blocks during updates
		if err := tx.Where("trip_id = ?", trip.ID).Delete(&entity.TripPlan{}).Error; err != nil {
			return err
		}
		if err := tx.Where("trip_id = ?", trip.ID).Delete(&entity.TripPricingTier{}).Error; err != nil {
			return err
		}
		return tx.Save(trip).Error
	})
}

func (r *tripRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Trip{}, id).Error
}

func (r *tripRepository) SearchSimilarTrips(ctx context.Context,destination string) ([]entity.Trip, error) {

    var trips []entity.Trip

    err := r.db.WithContext(ctx).
        Preload("Plans").
        Preload("PricingTiers").
        Where("LOWER(\"to\") LIKE LOWER(?)", "%"+destination+"%").
        Limit(5).
        Find(&trips).Error

    return trips, err
}
