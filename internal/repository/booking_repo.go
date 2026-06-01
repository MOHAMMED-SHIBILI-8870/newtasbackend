package repository

import (
	"backend/internal/entity"
	"context"

	"gorm.io/gorm"
)

type BookingRepository interface {
	CreateBooking(ctx context.Context, booking *entity.Booking) error
	CreateBookingTx(tx *gorm.DB, booking *entity.Booking) error
	GetBookingsByUserID(ctx context.Context, userID uint) ([]entity.Booking, error)
	GetBookingByID(ctx context.Context, id uint) (*entity.Booking, error)
	UpdateBookingPlans(ctx context.Context, bookingID uint, plans []entity.BookingPlan) error
	GetAllOrders(ctx context.Context) ([]entity.Booking, error)
}

type bookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) CreateBooking(ctx context.Context, booking *entity.Booking) error {
	return r.db.WithContext(ctx).Create(booking).Error
}

func (r *bookingRepository) CreateBookingTx(tx *gorm.DB, booking *entity.Booking) error {
	return tx.Create(booking).Error
}

func (r *bookingRepository) GetBookingsByUserID(ctx context.Context, userID uint) ([]entity.Booking, error) {

	var bookings []entity.Booking

	err := r.db.WithContext(ctx).
		Preload("Trip").
		Preload("CustomPlans").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&bookings).Error

	return bookings, err
}

func (r *bookingRepository) GetBookingByID(ctx context.Context, id uint) (*entity.Booking, error) {

	var booking entity.Booking

	err := r.db.WithContext(ctx).
		Preload("Trip").
		Preload("CustomPlans").
		First(&booking, id).Error

	return &booking, err
}

func (r *bookingRepository) UpdateBookingPlans(ctx context.Context, bookingID uint, plans []entity.BookingPlan) error {

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := tx.Where("booking_id = ?", bookingID).
			Delete(&entity.BookingPlan{}).Error; err != nil {
			return err
		}

		for i := range plans {
			plans[i].BookingID = bookingID
			plans[i].ID = 0
		}

		if len(plans) > 0 {
			return tx.Create(&plans).Error
		}

		return nil
	})
}

func (r *bookingRepository) GetAllOrders(ctx context.Context) ([]entity.Booking, error) {

	var bookings []entity.Booking

	err := r.db.WithContext(ctx).
		Preload("Trip").
		Preload("CustomPlans").
		Order("created_at DESC").
		Find(&bookings).Error

	return bookings, err
}
