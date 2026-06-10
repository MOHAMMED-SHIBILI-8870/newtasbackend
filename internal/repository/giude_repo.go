package repository

import (
	"backend/internal/entity"
	"context"

	"gorm.io/gorm"
)

type GuideRepository interface {
	GetProfile(ctx context.Context, userID uint) (*entity.Guide, error)
	UpdateProfile(ctx context.Context, guide *entity.Guide) error
}

type guideRepository struct {
	db *gorm.DB
}

func NewGuideRepository(db *gorm.DB) GuideRepository {
	return &guideRepository{
		db: db,
	}
}

func (r *guideRepository) GetProfile(
	ctx context.Context,
	userID uint,
) (*entity.Guide, error) {

	var guide entity.Guide

	err := r.db.
		WithContext(ctx).
		Where("user_id = ?", userID).
		First(&guide).Error

	if err != nil {
		return nil, err
	}

	return &guide, nil
}

func (r *guideRepository) UpdateProfile(
	ctx context.Context,
	guide *entity.Guide,
) error {

	return r.db.
		WithContext(ctx).
		Save(guide).
		Error
}