package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"errors"
	"log"
	"math"
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
	slotRepo            repository.TripSlotRepository
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
	slotRepos ...repository.TripSlotRepository,
) *BookingUsecase {
	var slotRepo repository.TripSlotRepository
	if len(slotRepos) > 0 {
		slotRepo = slotRepos[0]
	}

	return &BookingUsecase{
		bookingRepo:         br,
		tripRepo:            tr,
		userRepo:            ur,
		offerRepo:           offerRepo,
		slotRepo:            slotRepo,
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
	startDate *time.Time,
	endDate *time.Time,
) (*entity.Booking, error) {
	if u == nil || u.db == nil || u.bookingRepo == nil || u.tripRepo == nil || u.userRepo == nil {
		return nil, errors.New("booking service unavailable")
	}

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

	checkStart := trip.StartDate
	if startDate != nil {
		checkStart = *startDate
	}
	checkEnd := trip.EndDate
	if endDate != nil {
		checkEnd = *endDate
	}

	// CHECK FOR OVERLAPPING BOOKINGS
	overlap, err := u.bookingRepo.HasTripOverlap(
		ctx,
		userID,
		checkStart,
		checkEnd,
	)
	if err != nil {
		return nil, err
	}

	if overlap {
		return nil, errors.New(
			"overlap: you already have another trip booked during this period",
		)
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
		if offer.MaxUsage > 0 && offer.CurrentUsage >= offer.MaxUsage {
			return nil, errors.New("offer usage limit reached")
		}
		if offer.TripID != nil && *offer.TripID != tripID {
			return nil, errors.New("offer is not applicable for this trip")
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

		// All monetary math is rounded through minor units so float inputs do not
		// accumulate fractional-cent drift across booking and reporting flows.
		baseAmountMinor := moneyToMinorUnits(trip.Price) * int64(seats)
		baseAmount := moneyFromMinorUnits(baseAmountMinor)
		discountPercent := 0.0
		offerID := (*uint)(nil)
		coupon := ""
		finalAmountMinor := baseAmountMinor
		if offer != nil {
			offerID = &offer.ID
			coupon = offer.Code
			if offer.DiscountType == "fixed" {
				finalAmountMinor -= moneyToMinorUnits(offer.FixedDiscount)
			} else {
				discountPercent = roundMoney(offer.DiscountPercent)
				discountMinor := int64(math.Round(float64(baseAmountMinor) * discountPercent / 100))
				finalAmountMinor -= discountMinor
			}
			if finalAmountMinor < 0 {
				finalAmountMinor = 0
			}

			// Increment offer usage
			tx.Model(&entity.Offer{}).Where("id = ?", offer.ID).Update("current_usage", gorm.Expr("current_usage + ?", 1))
		}
		finalAmount := moneyFromMinorUnits(finalAmountMinor)

		advancePercent := 20.0

		advanceAmount := roundMoney(
			finalAmount * advancePercent / 100,
		)

		balanceAmount := roundMoney(
			finalAmount - advanceAmount,
		)

		booking := &entity.Booking{
			UserID:    userID,
			TripID:    tripID,
			VehicleID: vehicleID,
			OfferID:   offerID,

			StartDate: startDate,
			EndDate:   endDate,

			Status: "pending_payment",

			SeatsBooked:     seats,
			CouponCode:      coupon,
			DiscountPercent: discountPercent,

			BaseAmount:  baseAmount,
			FinalAmount: finalAmount,

			AdvancePercent: advancePercent,
			AdvanceAmount:  advanceAmount,
			BalanceAmount:  balanceAmount,
			PaymentStatus:  "pending",
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

	if u.notificationUsecase != nil && u.userRepo != nil {
		user, err := u.userRepo.GetByID(ctx, userID)
		if err == nil && user != nil {
			admins, err := u.userRepo.GetUsers(ctx,0 ,"admin", "")
			if err == nil {
				for _, admin := range admins {
					if err := u.notificationUsecase.CreateAdminBookingNotification(ctx, admin.ID, createdBooking.ID, "New booking from "+user.Email); err != nil {
						log.Printf("admin notification failed: %v", err)
					}
				}
			}
		}
	}

	return createdBooking, nil
}

func (u *BookingUsecase) BookSlot(
	ctx context.Context,
	slotID uint,
	userID uint,
	seats int,
	couponCode string,
	bookingType string,
	startDate *time.Time,
	endDate *time.Time,
) (*entity.Booking, error) {
	if u == nil || u.db == nil || u.bookingRepo == nil || u.tripRepo == nil || u.userRepo == nil || u.slotRepo == nil {
		return nil, errors.New("booking service unavailable")
	}

	if slotID == 0 {
		return nil, errors.New("slot id is required")
	}

	if seats <= 0 {
		seats = 1
	}

	bookingType, err := normalizeBookingType(bookingType)
	if err != nil {
		return nil, err
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
		if offer.MaxUsage > 0 && offer.CurrentUsage >= offer.MaxUsage {
			return nil, errors.New("offer usage limit reached")
		}
	}

	var createdBooking *entity.Booking
	err = u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var slot entity.TripSlot
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Trip").
			Preload("Trip.Plans").
			Preload("Vehicle").
			First(&slot, slotID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("slot not found")
			}
			return err
		}

		switch slot.Status {
		case "scheduled", "active":
		default:
			return errors.New("slot is not available for booking")
		}

		if offer != nil && offer.TripID != nil && *offer.TripID != slot.TripID {
			return errors.New("offer is not applicable for this trip")
		}

		checkStart := slot.StartDate
		if startDate != nil {
			checkStart = *startDate
		}
		checkEnd := slot.EndDate
		if endDate != nil {
			checkEnd = *endDate
		}

		// Prevent overlapping trip bookings
		overlap, err := u.bookingRepo.HasOverlappingBooking(
			ctx,
			userID,
			checkStart,
			checkEnd,
		)

		if err != nil {
			return err
		}

		if overlap {
			return errors.New(
				"overlap: you already have another trip booked during this period",
			)
		}

		if bookingType == "private" {
			if slot.BookedSeats > 0 {
				return errors.New("private booking requires an empty slot")
			}
			seats = slot.TotalSeats
		}

		if slot.AvailableSeats < seats {
			return errors.New("insufficient available seats")
		}

		var existingBooking entity.Booking
		if err := tx.Where("user_id = ? AND slot_id = ?", userID, slot.ID).First(&existingBooking).Error; err == nil {
			return errors.New("already booked this trip slot")
		}

		// All monetary math is rounded through minor units so float inputs do not
		// accumulate fractional-cent drift across booking and reporting flows.
		price := slot.PriceOverride
		if price <= 0 {
			price = slot.Trip.Price
		}
		baseAmountMinor := moneyToMinorUnits(price) * int64(seats)
		baseAmount := moneyFromMinorUnits(baseAmountMinor)
		discountPercent := 0.0
		offerID := (*uint)(nil)
		coupon := ""
		finalAmountMinor := baseAmountMinor
		if offer != nil {
			offerID = &offer.ID
			coupon = offer.Code
			if offer.DiscountType == "fixed" {
				finalAmountMinor -= moneyToMinorUnits(offer.FixedDiscount)
			} else {
				discountPercent = roundMoney(offer.DiscountPercent)
				discountMinor := int64(math.Round(float64(baseAmountMinor) * discountPercent / 100))
				finalAmountMinor -= discountMinor
			}
			if finalAmountMinor < 0 {
				finalAmountMinor = 0
			}

			// Increment offer usage
			tx.Model(&entity.Offer{}).Where("id = ?", offer.ID).Update("current_usage", gorm.Expr("current_usage + ?", 1))
		}
		finalAmount := moneyFromMinorUnits(finalAmountMinor)

		updatedAvailable := slot.AvailableSeats - seats
		updatedBooked := slot.BookedSeats + seats
		if bookingType == "private" {
			updatedAvailable = 0
			updatedBooked = slot.TotalSeats
		}

		slotUpdate := map[string]any{
			"available_seats": updatedAvailable,
			"booked_seats":    updatedBooked,
		}
		if updatedAvailable == 0 {
			slotUpdate["status"] = "fully_booked"
		}

		if result := tx.Model(&entity.TripSlot{}).
			Where("id = ? AND available_seats >= ?", slot.ID, seats).
			Updates(slotUpdate); result.Error != nil {
			return result.Error
		} else if result.RowsAffected == 0 {
			return errors.New("insufficient available seats")
		}

		advancePercent := 20.0

		advanceAmount := roundMoney(
			finalAmount * advancePercent / 100,
		)

		balanceAmount := roundMoney(
			finalAmount - advanceAmount,
		)

		booking := &entity.Booking{
			UserID:    userID,
			TripID:    slot.TripID,
			SlotID:    &slot.ID,
			VehicleID: slot.VehicleID,
			OfferID:   offerID,

			StartDate: startDate,
			EndDate:   endDate,

			BookingType: bookingType,
			Status:      "pending_payment",

			SeatsBooked:     seats,
			CouponCode:      coupon,
			DiscountPercent: discountPercent,

			BaseAmount:  baseAmount,
			FinalAmount: finalAmount,

			AdvancePercent: advancePercent,
			AdvanceAmount:  advanceAmount,
			BalanceAmount:  balanceAmount,
			PaymentStatus:  "pending",
		}

		for _, p := range slot.Trip.Plans {
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

		booking.Trip = slot.Trip
		booking.Slot = &slot
		createdBooking = booking
		return nil
	})
	if err != nil {
		return nil, err
	}

	if u.notificationUsecase != nil {
		if err := u.notificationUsecase.CreateBookingNotification(
			ctx,
			userID,
			createdBooking.ID,
			"Your slot booking has been created",
		); err != nil {
			log.Printf("user notification failed: %v", err)
		}
	}

	if u.notificationUsecase != nil && u.userRepo != nil {
		user, err := u.userRepo.GetByID(ctx, userID)
		if err == nil && user != nil {
			admins, err := u.userRepo.GetUsers(ctx, userID,"admin", "")
			if err == nil {
				for _, admin := range admins {
					if err := u.notificationUsecase.CreateAdminBookingNotification(ctx, admin.ID, createdBooking.ID, "New slot booking from "+user.Email); err != nil {
						log.Printf("admin notification failed: %v", err)
					}
				}
			}
		}
	}

	return createdBooking, nil
}

func (u *BookingUsecase) GetUserBookings(
	ctx context.Context,
	userID uint,
) ([]entity.Booking, error) {
	if u == nil || u.bookingRepo == nil {
		return nil, errors.New("booking service unavailable")
	}

	return u.bookingRepo.GetBookingsByUserID(
		ctx,
		userID,
	)
}
// Add this method to your existing BookingUsecase struct in internal/usecase/booking.go

func (u *BookingUsecase) GetBookingByID(
	ctx context.Context,
	bookingID uint,
	callerID uint,
	callerRole string,
) (*entity.Booking, error) {
	if u == nil || u.bookingRepo == nil {
		return nil, errors.New("booking service unavailable")
	}

	if bookingID == 0 {
		return nil, errors.New("invalid booking id lookup")
	}

	booking, err := u.bookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		// GORM returns gorm.ErrRecordNotFound if the ID doesn't exist in PostgreSQL
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("booking record not found")
		}
		return nil, err
	}

	// Ownership validation: only the booking owner or an admin may view
	if callerRole != "admin" && booking.UserID != callerID {
		return nil, errors.New("access denied")
	}

	return booking, nil
}

func (u *BookingUsecase) UpdateUserBookingPlans(
	ctx context.Context,
	bookingID uint,
	userID uint,
	plans []entity.BookingPlan,
) error {
	if u == nil || u.bookingRepo == nil {
		return errors.New("booking service unavailable")
	}

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
	if u == nil || u.bookingRepo == nil {
		return nil, errors.New("booking service unavailable")
	}

	return u.bookingRepo.GetAllOrders(ctx)
}

func moneyToMinorUnits(amount float64) int64 {
	return int64(math.Round(amount * 100))
}

func moneyFromMinorUnits(minor int64) float64 {
	return float64(minor) / 100
}

func roundMoney(amount float64) float64 {
	return math.Round(amount*100) / 100
}
