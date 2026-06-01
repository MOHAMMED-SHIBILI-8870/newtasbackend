package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BookingUsecase struct {
	bookingRepo         repository.BookingRepository
	tripRepo            repository.TripRepository
	userRepo            repository.UserRepository
	offerRepo           repository.OfferRepository
	db                  *gorm.DB
	notificationUsecase *NotificationUsecase
}

func NewBookingUsecase(
	br repository.BookingRepository,
	tr repository.TripRepository,
	ur repository.UserRepository,
	offerRepo repository.OfferRepository,
	db *gorm.DB,
	notificationUsecase *NotificationUsecase,
) *BookingUsecase {

	return &BookingUsecase{
		bookingRepo:         br,
		tripRepo:            tr,
		userRepo:            ur,
		offerRepo:           offerRepo,
		db:                  db,
		notificationUsecase: notificationUsecase,
	}
}

func (u *BookingUsecase) BookTrip(
	ctx context.Context,
	tripID uint,
	userID uint,
	seats int,
	couponCode string,
) (*entity.Booking, error) {

	if seats <= 0 {
		seats = 1
	}

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

	var offer *entity.Offer
	if strings.TrimSpace(couponCode) != "" && u.offerRepo != nil {
		offer, err = u.offerRepo.GetByCode(ctx, couponCode)
		if err != nil {
			return nil, err
		}
		if offer == nil {
			return nil, errors.New("offer not found")
		}
		if !offer.Active {
			return nil, errors.New("offer is inactive")
		}
		if offer.ExpiryDate.Before(time.Now()) {
			return nil, errors.New("offer has expired")
		}
	}

	var createdBooking *entity.Booking
	err = u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var vehicle entity.Vehicle
		vehicleFound := false
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("trip_id = ?", tripID).
			First(&vehicle).Error; err == nil {
			vehicleFound = true
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var vehicleID *uint
		if vehicleFound {
			if vehicle.AvailableSeats < seats {
				return errors.New("insufficient available seats")
			}

			result := tx.Model(&entity.Vehicle{}).
				Where("id = ? AND available_seats >= ?", vehicle.ID, seats).
				Update("available_seats", gorm.Expr("available_seats - ?", seats))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return errors.New("insufficient available seats")
			}

			vehicleID = &vehicle.ID
		}

		baseAmount := trip.Price * float64(seats)
		discountPercent := 0.0
		offerID := (*uint)(nil)
		coupon := ""
		finalAmount := baseAmount
		if offer != nil {
			discountPercent = offer.DiscountPercent
			offerID = &offer.ID
			coupon = offer.Code
			finalAmount = baseAmount - (baseAmount * discountPercent / 100)
			if finalAmount < 0 {
				finalAmount = 0
			}
		}

		booking := &entity.Booking{
			UserID:          userID,
			TripID:          tripID,
			VehicleID:       vehicleID,
			OfferID:         offerID,
			Status:          "confirmed",
			SeatsBooked:     seats,
			CouponCode:      coupon,
			DiscountPercent: discountPercent,
			BaseAmount:      baseAmount,
			FinalAmount:     finalAmount,
		}

		for _, p := range trip.Plans {
			booking.CustomPlans = append(
				booking.CustomPlans,
				entity.BookingPlan{
					DayNumber:   p.DayNumber,
					Title:       p.Title,
					Description: p.Description,
					Location:    p.Location,
					StartTime:   p.StartTime,
					EndTime:     p.EndTime,
					Category:    p.Category,
					Cost:        p.Cost,
				},
			)
		}

		if err := u.bookingRepo.CreateBookingTx(tx, booking); err != nil {
			return err
		}

		createdBooking = booking
		return nil
	})
	if err != nil {
		return nil, err
	}

	// =====================================
	// USER NOTIFICATION
	// =====================================

	if u.notificationUsecase != nil {

		err := u.notificationUsecase.CreateBookingNotification(
			ctx,
			userID,
			createdBooking.ID,
			"Your booking has been confirmed",
		)

		if err != nil {

			log.Printf(
				"user notification failed: %v",
				err,
			)
		}
	}

	// =====================================
	// ADMIN NOTIFICATION
	// =====================================

	user, err := u.userRepo.GetByID(
		ctx,
		userID,
	)

	if err == nil {

		adminNotification := &entity.Notification{
			Type:      "booking",
			Title:     "New Booking",
			Message:   "New booking from " + user.Email,
			BookingID: &createdBooking.ID,
			IsRead:    false,
			IsAdmin:   true,
		}

		err = u.notificationUsecase.CreateNotification(
			ctx,
			adminNotification,
		)

		if err != nil {

			log.Printf(
				"admin notification failed: %v",
				err,
			)
		}
	}

	return createdBooking, nil
}

func (u *BookingUsecase) GetUserBookings(
	ctx context.Context,
	userID uint,
) ([]entity.Booking, error) {

	return u.bookingRepo.GetBookingsByUserID(
		ctx,
		userID,
	)
}

func (u *BookingUsecase) UpdateUserBookingPlans(
	ctx context.Context,
	bookingID uint,
	userID uint,
	plans []entity.BookingPlan,
) error {

	booking, err := u.bookingRepo.GetBookingByID(
		ctx,
		bookingID,
	)

	if err != nil {
		return errors.New("booking not found")
	}

	if booking.UserID != userID {
		return errors.New("access denied")
	}

	return u.bookingRepo.UpdateBookingPlans(
		ctx,
		bookingID,
		plans,
	)
}

func (u *BookingUsecase) GetAllOrders(
	ctx context.Context,
) ([]entity.Booking, error) {

	return u.bookingRepo.GetAllOrders(ctx)
}
