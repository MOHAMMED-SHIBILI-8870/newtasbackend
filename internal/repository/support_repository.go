package repository

import (
	"backend/internal/entity"
	"context"

	"gorm.io/gorm"
)

type SupportRepository interface {
	CreateRequest(ctx context.Context, req *entity.SupportRequest) error
	GetRequestByID(ctx context.Context, id uint) (*entity.SupportRequest, error)
	ListRequests(ctx context.Context, userID *uint, status string) ([]entity.SupportRequest, error)
	UpdateRequest(ctx context.Context, req *entity.SupportRequest) error
}

type supportRepository struct {
	db *gorm.DB
}

func NewSupportRepository(db *gorm.DB) SupportRepository {
	return &supportRepository{db: db}
}

func (r *supportRepository) CreateRequest(ctx context.Context, req *entity.SupportRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *supportRepository) GetRequestByID(ctx context.Context, id uint) (*entity.SupportRequest, error) {
	var req entity.SupportRequest
	err := r.db.WithContext(ctx).Preload("User").Preload("Agent").First(&req, id).Error
	return &req, err
}

func (r *supportRepository) ListRequests(ctx context.Context, userID *uint, status string) ([]entity.SupportRequest, error) {
	var reqs []entity.SupportRequest
	query := r.db.WithContext(ctx).Preload("User").Preload("Agent").Order("created_at DESC")
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Find(&reqs).Error
	return reqs, err
}

func (r *supportRepository) UpdateRequest(ctx context.Context, req *entity.SupportRequest) error {
	return r.db.WithContext(ctx).Save(req).Error
}
