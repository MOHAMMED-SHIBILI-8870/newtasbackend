package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type ReviewUsecase struct {
	reviewRepo  repository.ReviewRepository
	bookingRepo repository.BookingRepository
	tripRepo    repository.TripRepository
}

func NewReviewUsecase(
	reviewRepo repository.ReviewRepository,
	bookingRepo repository.BookingRepository,
	tripRepo repository.TripRepository,
) *ReviewUsecase {
	return &ReviewUsecase{
		reviewRepo:  reviewRepo,
		bookingRepo: bookingRepo,
		tripRepo:    tripRepo,
	}
}

func (u *ReviewUsecase) CreateReview(ctx context.Context, userID uint, tripID uint, rating int, comment string) (*entity.Review, error) {
	if userID == 0 {
		return nil, errors.New("user id is required")
	}
	if tripID == 0 {
		return nil, errors.New("trip id is required")
	}
	if rating < 1 || rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}

	trip, err := u.tripRepo.GetByID(ctx, tripID)
	if err != nil {
		return nil, err
	}
	if trip == nil {
		return nil, fmt.Errorf("trip not found")
	}
	if time.Now().Before(trip.EndDate) {
		return nil, errors.New("trip has not been completed yet")
	}

	bookings, err := u.bookingRepo.GetBookingsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	booked := false
	for _, booking := range bookings {
		if booking.TripID == tripID && booking.Status != "cancelled" {
			booked = true
			break
		}
	}
	if !booked {
		return nil, errors.New("booking not found for this trip")
	}

	existingReviews, err := u.reviewRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, review := range existingReviews {
		if review.TripID == tripID {
			return nil, errors.New("review already exists for this trip")
		}
	}

	review := &entity.Review{
		UserID:  userID,
		TripID:  tripID,
		Rating:  rating,
		Comment: strings.TrimSpace(comment),
	}

	if err := u.reviewRepo.Create(ctx, review); err != nil {
		return nil, err
	}

	return review, nil
}

func (u *ReviewUsecase) ListTripReviews(ctx context.Context, tripID uint) ([]entity.Review, error) {
	if tripID == 0 {
		return nil, errors.New("trip id is required")
	}
	return u.reviewRepo.GetByTripID(ctx, tripID)
}

func (u *ReviewUsecase) ListUserReviews(ctx context.Context, userID uint) ([]entity.Review, error) {
	if userID == 0 {
		return nil, errors.New("user id is required")
	}
	return u.reviewRepo.GetByUserID(ctx, userID)
}

func (u *ReviewUsecase) ListAllReviews(ctx context.Context) ([]entity.Review, error) {
	return u.reviewRepo.GetAll(ctx)
}

func (u *ReviewUsecase) GetTripAverageRating(ctx context.Context, tripID uint) (float64, int64, error) {
	reviews, err := u.reviewRepo.GetByTripID(ctx, tripID)
	if err != nil {
		return 0, 0, err
	}

	if len(reviews) == 0 {
		return 0, 0, nil
	}

	var sum int
	for _, review := range reviews {
		sum += review.Rating
	}

	average := float64(sum) / float64(len(reviews))
	return average, int64(len(reviews)), nil
}

func (u *ReviewUsecase) DeleteReview(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("review id is required")
	}
	return u.reviewRepo.Delete(ctx, id)
}
