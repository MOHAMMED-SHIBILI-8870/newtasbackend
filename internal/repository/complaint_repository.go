package repository

import (
	"backend/internal/entity"
	"context"

	"gorm.io/gorm"
)

type ComplaintRepository interface {
	Create(ctx context.Context, complaint *entity.Complaint) error
	Update(ctx context.Context, complaint *entity.Complaint) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*entity.Complaint, error)
	GetByUserID(ctx context.Context, userID uint) ([]entity.Complaint, error)
	GetByBookingID(ctx context.Context, bookingID uint) ([]entity.Complaint, error)
	GetAll(ctx context.Context) ([]entity.Complaint, error)
}

type complaintRepository struct {
	db *gorm.DB
}

func NewComplaintRepository(db *gorm.DB) ComplaintRepository {
	return &complaintRepository{db: db}
}

func (r *complaintRepository) Create(ctx context.Context, complaint *entity.Complaint) error {
	return r.db.WithContext(ctx).Create(complaint).Error
}

func (r *complaintRepository) Update(ctx context.Context, complaint *entity.Complaint) error {
	return r.db.WithContext(ctx).Save(complaint).Error
}

func (r *complaintRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Complaint{}, id).Error
}

func (r *complaintRepository) GetByID(ctx context.Context, id uint) (*entity.Complaint, error) {
	var complaint entity.Complaint
	err := r.db.WithContext(ctx).Preload("User").First(&complaint, id).Error
	if err != nil {
		return nil, err
	}
	return &complaint, nil
}

func (r *complaintRepository) GetByUserID(ctx context.Context, userID uint) ([]entity.Complaint, error) {
	var complaints []entity.Complaint
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&complaints).Error
	return complaints, err
}

func (r *complaintRepository) GetByBookingID(ctx context.Context, bookingID uint) ([]entity.Complaint, error) {
	var complaints []entity.Complaint
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("booking_id = ?", bookingID).
		Order("created_at DESC").
		Find(&complaints).Error
	return complaints, err
}

func (r *complaintRepository) GetAll(ctx context.Context) ([]entity.Complaint, error) {
	var complaints []entity.Complaint
	err := r.db.WithContext(ctx).Preload("User").Order("created_at DESC").Find(&complaints).Error
	return complaints, err
}
