package repository

import (
	"backend/internal/entity"
	"gorm.io/gorm"
)

type VerificationRepository interface {
	Create(verification *entity.Verification) error
	GetByID(id uint) (*entity.Verification, error)
	GetByBookingID(bookingID uint) (*entity.Verification, error)
}

type verificationRepository struct {
	db *gorm.DB
}

func NewVerificationRepository(db *gorm.DB) VerificationRepository {
	return &verificationRepository{db: db}
}

func (r *verificationRepository) Create(verification *entity.Verification) error {
	return r.db.Create(verification).Error
}

func (r *verificationRepository) GetByID(id uint) (*entity.Verification, error) {
	var verification entity.Verification
	err := r.db.First(&verification, id).Error
	if err != nil {
		return nil, err
	}
	return &verification, nil
}

func (r *verificationRepository) GetByBookingID(bookingID uint) (*entity.Verification, error) {
	var verification entity.Verification
	err := r.db.Where("booking_id = ?", bookingID).First(&verification).Error
	if err != nil {
		return nil, err
	}
	return &verification, nil
}
