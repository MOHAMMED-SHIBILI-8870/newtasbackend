package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReviewUsecase struct {
	reviewRepo  repository.ReviewRepository
	bookingRepo repository.BookingRepository
	tripRepo    repository.TripRepository
	db          *gorm.DB
}

func NewReviewUsecase(
	reviewRepo repository.ReviewRepository,
	bookingRepo repository.BookingRepository,
	tripRepo repository.TripRepository,
	db *gorm.DB,
) *ReviewUsecase {
	return &ReviewUsecase{
		reviewRepo:  reviewRepo,
		bookingRepo: bookingRepo,
		tripRepo:    tripRepo,
		db:          db,
	}
}

func (u *ReviewUsecase) CreateReview(ctx context.Context, userID uint, tripID uint, guideID *uint, rating int, comment string) (*entity.Review, error) {
	if userID == 0 {
		return nil, errors.New("user id is required")
	}
	if tripID == 0 {
		return nil, errors.New("trip id is required")
	}
	if rating < 1 || rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}

	if u.db == nil {
		return nil, errors.New("database unavailable")
	}

	review := &entity.Review{
		UserID:  userID,
		TripID:  tripID,
		GuideID: guideID,
		Rating:  rating,
		Comment: strings.TrimSpace(comment),
	}

	if err := u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var trip entity.Trip
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Plans").
			First(&trip, tripID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("trip not found")
			}
			return err
		}

		var booking entity.Booking
		if err := tx.Model(&entity.Booking{}).
			Where("user_id = ? AND trip_id = ? AND status <> ?", userID, tripID, "cancelled").
			Order("end_date DESC").
			First(&booking).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("you can only review trips that you have booked")
			}
			return err
		}

		var existing entity.Review
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND trip_id = ?", userID, tripID).
			First(&existing).Error
		if err == nil {
			return errors.New("review already exists for this trip")
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		return tx.Create(review).Error
	}); err != nil {
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
