package repository

import (
	"backend/internal/entity"
	"context"

	"gorm.io/gorm"
)

type ReviewRepository interface {
	Create(ctx context.Context, review *entity.Review) error
	Update(ctx context.Context, review *entity.Review) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*entity.Review, error)
	GetByTripID(ctx context.Context, tripID uint) ([]entity.Review, error)
	GetByUserID(ctx context.Context, userID uint) ([]entity.Review, error)
	GetAll(ctx context.Context) ([]entity.Review, error)
}

type reviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(ctx context.Context, review *entity.Review) error {
	return r.db.WithContext(ctx).Create(review).Error
}

func (r *reviewRepository) Update(ctx context.Context, review *entity.Review) error {
	return r.db.WithContext(ctx).Save(review).Error
}

func (r *reviewRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Review{}, id).Error
}

func (r *reviewRepository) GetByID(ctx context.Context, id uint) (*entity.Review, error) {
	var review entity.Review
	err := r.db.WithContext(ctx).First(&review, id).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *reviewRepository) GetByTripID(ctx context.Context, tripID uint) ([]entity.Review, error) {
	var reviews []entity.Review
	err := r.db.WithContext(ctx).
		Preload("User").Preload("Trip").
		Where("trip_id = ?", tripID).
		Order("created_at DESC").
		Find(&reviews).Error
	return reviews, err
}

func (r *reviewRepository) GetByUserID(ctx context.Context, userID uint) ([]entity.Review, error) {
	var reviews []entity.Review
	err := r.db.WithContext(ctx).
		Preload("User").Preload("Trip").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&reviews).Error
	return reviews, err
}

func (r *reviewRepository) GetAll(ctx context.Context) ([]entity.Review, error) {
	var reviews []entity.Review
	err := r.db.WithContext(ctx).
		Preload("User").Preload("Trip").
		Order("created_at DESC").
		Find(&reviews).Error
	return reviews, err
}
