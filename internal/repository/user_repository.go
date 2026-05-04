package repository

import (
	"backend/internal/entity"
	"context"

	"gorm.io/gorm"
)

// TripRepository defines the interface for database operations
type TripRepository interface {
	Create(ctx context.Context, trip *entity.Trip) error
	GetByID(ctx context.Context, id uint) (*entity.Trip, error)
	GetByUserID(ctx context.Context, userID uint) ([]entity.Trip, error)
	Update(ctx context.Context, trip *entity.Trip) error
	Delete(ctx context.Context, id uint) error
}

type tripRepository struct {
	db *gorm.DB
}

// NewTripRepository creates a new instance of the repository
func NewTripRepository(db *gorm.DB) TripRepository {
	return &tripRepository{db: db}
}

// Create inserts a new trip into the database
func (r *tripRepository) Create(ctx context.Context, trip *entity.Trip) error {
	return r.db.WithContext(ctx).Create(trip).Error
}

// GetByID finds a single trip and preloads the owner (User) details
func (r *tripRepository) GetByID(ctx context.Context, id uint) (*entity.Trip, error) {
	var trip entity.Trip
	err := r.db.WithContext(ctx).Preload("User").First(&trip, id).Error
	return &trip, err
}

// GetByUserID fetches all trips belonging to a specific user
func (r *tripRepository) GetByUserID(ctx context.Context, userID uint) ([]entity.Trip, error) {
	var trips []entity.Trip
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&trips).Error
	return trips, err
}

// Update saves changes to an existing trip record
func (r *tripRepository) Update(ctx context.Context, trip *entity.Trip) error {
	return r.db.WithContext(ctx).Save(trip).Error
}

// Delete marks a trip as deleted (Soft Delete)
func (r *tripRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Trip{}, id).Error
}