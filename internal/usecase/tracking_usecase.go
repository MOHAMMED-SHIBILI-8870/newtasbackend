package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"errors"
	"fmt"
)

type TrackingUsecase struct {
	trackingRepo repository.TrackingRepository
	bookingRepo  repository.BookingRepository
	vehicleRepo  repository.VehicleRepository
}

func NewTrackingUsecase(
	trackingRepo repository.TrackingRepository,
	bookingRepo repository.BookingRepository,
	vehicleRepo repository.VehicleRepository,
) *TrackingUsecase {
	return &TrackingUsecase{
		trackingRepo: trackingRepo,
		bookingRepo:  bookingRepo,
		vehicleRepo:  vehicleRepo,
	}
}

func (u *TrackingUsecase) UpdateLocation(ctx context.Context, bookingID uint, vehicleID uint, latitude float64, longitude float64) (*entity.Tracking, error) {
	if bookingID == 0 || vehicleID == 0 {
		return nil, errors.New("booking and vehicle ids are required")
	}
	if latitude < -90 || latitude > 90 {
		return nil, errors.New("invalid latitude")
	}
	if longitude < -180 || longitude > 180 {
		return nil, errors.New("invalid longitude")
	}

	booking, err := u.bookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	vehicle, err := u.vehicleRepo.GetByID(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	if vehicle == nil {
		return nil, fmt.Errorf("vehicle not found")
	}
	if vehicle.TripID == nil {
		return nil, errors.New("vehicle is not assigned to a trip")
	}

	if *vehicle.TripID != booking.TripID {
		return nil, errors.New("vehicle is not assigned to this booking trip")
	}

	tracking := &entity.Tracking{
		BookingID: bookingID,
		VehicleID: vehicleID,
		Latitude:  latitude,
		Longitude: longitude,
	}

	if err := u.trackingRepo.Create(ctx, tracking); err != nil {
		return nil, err
	}

	return tracking, nil
}

func (u *TrackingUsecase) GetLatestByBookingID(ctx context.Context, bookingID uint) (*entity.Tracking, error) {
	if bookingID == 0 {
		return nil, errors.New("booking id is required")
	}
	return u.trackingRepo.GetLatestByBookingID(ctx, bookingID)
}

func (u *TrackingUsecase) GetLatestForUser(ctx context.Context, userID uint, bookingID uint, role string) (*entity.Tracking, error) {
	if bookingID == 0 {
		return nil, errors.New("booking id is required")
	}

	if NormalizeRole(role) == "admin" || NormalizeRole(role) == "driver" {
		return u.GetLatestByBookingID(ctx, bookingID)
	}

	booking, err := u.bookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}
	if booking.UserID != userID {
		return nil, errors.New("access denied")
	}

	return u.GetLatestByBookingID(ctx, bookingID)
}

func (u *TrackingUsecase) GetTrackingsByBookingID(ctx context.Context, bookingID uint) ([]entity.Tracking, error) {
	if bookingID == 0 {
		return nil, errors.New("booking id is required")
	}
	return u.trackingRepo.GetByBookingID(ctx, bookingID)
}

func (u *TrackingUsecase) GetTrackingsForUser(ctx context.Context, userID uint, bookingID uint, role string) ([]entity.Tracking, error) {
	if bookingID == 0 {
		return nil, errors.New("booking id is required")
	}

	if NormalizeRole(role) == "admin" || NormalizeRole(role) == "driver" {
		return u.GetTrackingsByBookingID(ctx, bookingID)
	}

	booking, err := u.bookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}
	if booking.UserID != userID {
		return nil, errors.New("access denied")
	}

	return u.GetTrackingsByBookingID(ctx, bookingID)
}

func (u *TrackingUsecase) GetAll(ctx context.Context) ([]entity.Tracking, error) {
	return u.trackingRepo.GetAll(ctx)
}
