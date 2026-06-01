package repository

import (
	"backend/internal/entity"
	"context"

	"gorm.io/gorm"
)

type AITripRequestRepository interface {
	Create(ctx context.Context, request *entity.AITripRequest) error
	GetByID(ctx context.Context, id uint) (*entity.AITripRequest, error)
	GetByUserID(ctx context.Context, userID uint) ([]entity.AITripRequest, error)
	GetAll(ctx context.Context) ([]entity.AITripRequest, error)
	Update(ctx context.Context, request *entity.AITripRequest) error
}

type aiTripRequestRepository struct {
	db *gorm.DB
}

func NewAITripRequestRepository(db *gorm.DB) AITripRequestRepository {
	return &aiTripRequestRepository{db: db}
}

func (r *aiTripRequestRepository) Create(ctx context.Context, request *entity.AITripRequest) error {
	return r.db.WithContext(ctx).Create(request).Error
}

func (r *aiTripRequestRepository) GetByID(ctx context.Context, id uint) (*entity.AITripRequest, error) {
	var request entity.AITripRequest
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Trip").
		Preload("ReviewedBy").
		First(&request, id).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *aiTripRequestRepository) GetByUserID(ctx context.Context, userID uint) ([]entity.AITripRequest, error) {
	var requests []entity.AITripRequest
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Trip").
		Preload("ReviewedBy").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&requests).Error
	return requests, err
}

func (r *aiTripRequestRepository) GetAll(ctx context.Context) ([]entity.AITripRequest, error) {
	var requests []entity.AITripRequest
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Trip").
		Preload("ReviewedBy").
		Order("created_at DESC").
		Find(&requests).Error
	return requests, err
}

func (r *aiTripRequestRepository) Update(ctx context.Context, request *entity.AITripRequest) error {
	return r.db.WithContext(ctx).Save(request).Error
}
