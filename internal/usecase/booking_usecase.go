package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"errors"
	"log"
)

type BookingUsecase struct {
	bookingRepo         repository.BookingRepository
	tripRepo            repository.TripRepository
	notificationUsecase *NotificationUsecase
}

func NewBookingUsecase(br repository.BookingRepository, tr repository.TripRepository, notificationUsecase *NotificationUsecase) *BookingUsecase {
	return &BookingUsecase{
		bookingRepo:         br,
		tripRepo:            tr,
		notificationUsecase: notificationUsecase,
	}
}

func (u *BookingUsecase) BookTrip(ctx context.Context, tripID uint, userID uint) (*entity.Booking, error) {
	trip, err := u.tripRepo.GetByID(ctx, tripID)
	if err != nil {
		return nil, errors.New("trip not found")
	}
	if trip == nil {
		return nil, errors.New("trip not found")
	}
	if trip.Status != "" && trip.Status != "active" {
		return nil, errors.New("trip is not available for booking")
	}

	booking := &entity.Booking{
		UserID: userID,
		TripID: tripID,
		Status: "confirmed",
	}

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

	if u.notificationUsecase != nil {
		if err := u.notificationUsecase.CreateBookingNotification(
			ctx,
			userID,
			booking.ID,
			"Your booking has been confirmed",
		); err != nil {
			log.Printf("booking notification creation failed for booking %d: %v", booking.ID, err)
		}
	}

	return booking, nil
}

func (u *BookingUsecase) GetUserBookings(ctx context.Context, userID uint) ([]entity.Booking, error) {
	return u.bookingRepo.GetBookingsByUserID(ctx, userID)
}

func (u *BookingUsecase) UpdateUserBookingPlans(ctx context.Context, bookingID uint, userID uint, plans []entity.BookingPlan) error {
	booking, err := u.bookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return errors.New("booking not found")
	}

	if booking.UserID != userID {
		return errors.New("access denied")
	}

	return u.bookingRepo.UpdateBookingPlans(ctx, bookingID, plans)
}

func (u *BookingUsecase) GetAllOrders(ctx context.Context) ([]entity.Booking, error) {
	return u.bookingRepo.GetAllOrders(ctx)
}