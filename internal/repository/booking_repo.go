package repository

import (
	"backend/internal/entity"
	"context"
	"time"

	"gorm.io/gorm"
)

type BookingRepository interface {
	CreateBooking(ctx context.Context, booking *entity.Booking) error
	CreateBookingTx(tx *gorm.DB, booking *entity.Booking) error
	UpdateBooking(ctx context.Context,booking *entity.Booking)error
	GetBookingsByUserID(ctx context.Context, userID uint) ([]entity.Booking, error)
	GetBookingByID(ctx context.Context, id uint) (*entity.Booking, error)
	UpdateBookingPlans(ctx context.Context, bookingID uint, plans []entity.BookingPlan) error
	GetAllOrders(ctx context.Context) ([]entity.Booking, error)
	HasOverlappingBooking(ctx context.Context, userID uint, startDate time.Time, endDate time.Time) (bool, error)
	HasTripOverlap(ctx context.Context,userID uint,startDate time.Time,endDate time.Time) (bool, error)
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

func (r *bookingRepository) UpdateBooking(ctx context.Context,booking *entity.Booking)error{
	return r.db.WithContext(ctx).Save(booking).Error
}

func (r *bookingRepository) CreateBookingTx(tx *gorm.DB, booking *entity.Booking) error {
	return tx.Create(booking).Error
}

func (r *bookingRepository) GetBookingsByUserID(ctx context.Context, userID uint) ([]entity.Booking, error) {

	var bookings []entity.Booking

	err := r.db.WithContext(ctx).
		Preload("Trip").
		Preload("Slot").
		Preload("Slot.Trip").
		Preload("Slot.Vehicle").
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
		Preload("Slot").
		Preload("Slot.Trip").
		Preload("Slot.Vehicle").
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
		Preload("Slot").
		Preload("Slot.Trip").
		Preload("Slot.Vehicle").
		Preload("CustomPlans").
		Order("created_at DESC").
		Find(&bookings).Error

	return bookings, err
}

func (r *bookingRepository) HasOverlappingBooking(ctx context.Context, userID uint, startDate time.Time, endDate time.Time) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).Model(&entity.Booking{}).
		Joins("JOIN trip_slots ts ON ts.id = bookings.slot_id").
		Where("bookings.user_id = ?", userID).
		Where("bookings.status IN ?", []string{
			"pending_payment",
			"confirmed",
		}).
		Where(
			"COALESCE(bookings.start_date, ts.start_date) <= ? AND COALESCE(bookings.end_date, ts.end_date) >= ?",
			endDate,
			startDate,
		).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *bookingRepository) HasTripOverlap(ctx context.Context,userID uint,startDate time.Time,endDate time.Time) (bool, error) {

	var count int64

	err := r.db.WithContext(ctx).
		Model(&entity.Booking{}).
		Joins("JOIN trips t ON t.id = bookings.trip_id").
		Where("bookings.user_id = ?", userID).
		Where("bookings.status IN ?", []string{
			"confirmed",
			"pending_payment",
		}).
		Where(
			"COALESCE(bookings.start_date, t.start_date) <= ? AND COALESCE(bookings.end_date, t.end_date) >= ?",
			endDate,
			startDate,
		).
		Count(&count).
		Error

	return count > 0, err
}