package repository

import (
	"backend/internal/entity"
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type OfferRepository interface {
	Create(ctx context.Context, offer *entity.Offer) error
	Update(ctx context.Context, offer *entity.Offer) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*entity.Offer, error)
	GetByCode(ctx context.Context, code string) (*entity.Offer, error)
	List(ctx context.Context) ([]entity.Offer, error)
	GetActive(ctx context.Context) ([]entity.Offer, error)
}

type offerRepository struct {
	db *gorm.DB
}

func NewOfferRepository(db *gorm.DB) OfferRepository {
	return &offerRepository{db: db}
}

func (r *offerRepository) Create(ctx context.Context, offer *entity.Offer) error {
	return r.db.WithContext(ctx).Create(offer).Error
}

func (r *offerRepository) Update(ctx context.Context, offer *entity.Offer) error {
	return r.db.WithContext(ctx).Save(offer).Error
}

func (r *offerRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Offer{}, id).Error
}

func (r *offerRepository) GetByID(ctx context.Context, id uint) (*entity.Offer, error) {
	var offer entity.Offer
	err := r.db.WithContext(ctx).First(&offer, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("offer not found")
		}
		return nil, err
	}
	return &offer, nil
}

func (r *offerRepository) GetByCode(ctx context.Context, code string) (*entity.Offer, error) {
	var offer entity.Offer
	err := r.db.WithContext(ctx).Where("LOWER(code) = LOWER(?)", code).First(&offer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &offer, nil
}

func (r *offerRepository) List(ctx context.Context) ([]entity.Offer, error) {
	var offers []entity.Offer
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&offers).Error
	return offers, err
}

func (r *offerRepository) GetActive(ctx context.Context) ([]entity.Offer, error) {
	var offers []entity.Offer
	err := r.db.WithContext(ctx).
		Where("active = ? AND expiry_date >= ?", true, time.Now()).
		Order("expiry_date ASC").
		Find(&offers).Error
	return offers, err
}
