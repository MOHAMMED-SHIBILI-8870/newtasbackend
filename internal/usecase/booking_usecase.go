package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"errors"
)

type BookingUsecase struct {
	bookingRepo repository.BookingRepository
	tripRepo    repository.TripRepository
}

func NewBookingUsecase(br repository.BookingRepository, tr repository.TripRepository) *BookingUsecase {
	return &BookingUsecase{
		bookingRepo: br,
		tripRepo:    tr,
	}
}

func (u *BookingUsecase) BookTrip(ctx context.Context, tripID uint, userID uint) (*entity.Booking, error) {
	// Retrieve the latest master plan state template built by admin panels
	trip, err := u.tripRepo.GetByID(ctx, tripID)
	if err != nil {
		return nil, errors.New("master travel package configuration reference not found")
	}

	booking := &entity.Booking{
		UserID: userID,
		TripID: tripID,
		Status: "confirmed",
	}

	// Copy the static trip design parameters natively onto user custom blocks
	for _, p := range trip.Plans {
		booking.CustomPlans = append(booking.CustomPlans, entity.BookingPlan{
			DayNumber:   p.DayNumber,
			Title:       p.Title,
			Description: p.Description,
			Location:    p.Location,
			StartTime:   p.StartTime,
			EndTime:     p.EndTime,
			Category:    p.Category,
			Cost:        p.Cost,
		})
	}

	if err := u.bookingRepo.CreateBooking(ctx, booking); err != nil {
		return nil, err
	}
	return booking, nil
}

func (u *BookingUsecase) GetUserBookings(ctx context.Context, userID uint) ([]entity.Booking, error) {
	return u.bookingRepo.GetBookingsByUserID(ctx, userID)
}

func (u *BookingUsecase) UpdateUserBookingPlans(ctx context.Context, bookingID uint, userID uint, plans []entity.BookingPlan) error {
	booking, err := u.bookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return errors.New("targeted booking resource profile missing")
	}
	
	// Validation check: Verify the user owns the booking they are attempting to alter
	if booking.UserID != userID {
		return errors.New("action prohibited: identity context ownership verification failed")
	}

	return u.bookingRepo.UpdateBookingPlans(ctx, bookingID, plans)
}

func (u *BookingUsecase) GetAllOrders(ctx context.Context) ([]entity.Booking, error) {
	return u.bookingRepo.GetAllOrders(ctx)
}