package usecase

import (
	"context"
	"errors"
	"backend/internal/entity"
	"backend/internal/repository"
)

type TripUsecase struct {
	repo repository.TripRepository
}

func NewTripUsecase(r repository.TripRepository) *TripUsecase {
	return &TripUsecase{repo: r}
}

// CreateTrip handles validation before saving
func (u *TripUsecase) CreateTrip(ctx context.Context, trip *entity.Trip) error {
	if trip.Destination == "" {
		return errors.New("destination is required")
	}
	if trip.Budget < 0 {
		return errors.New("budget cannot be negative")
	}
	return u.repo.Create(ctx, trip)
}

// GetTripDetails fetches a single trip
func (u *TripUsecase) GetTripDetails(ctx context.Context, id uint) (*entity.Trip, error) {
	return u.repo.GetByID(ctx, id)
}

// GetTripsByOwner fetches all trips for a specific user
func (u *TripUsecase) GetTripsByOwner(ctx context.Context, userID uint) ([]entity.Trip, error) {
	return u.repo.GetByUserID(ctx, userID)
}

// UpdateTrip checks if the user owns the trip before updating
func (u *TripUsecase) UpdateTrip(ctx context.Context, id uint, input entity.UpdateTripInput, userID uint) error {
	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// ownership check
	if existing.UserId != userID {
		return errors.New("you do not have permission to update this trip")
	}

	// update only provided fields
	if input.Destination != nil {
		existing.Destination = *input.Destination
	}
	if input.Budget != nil {
		existing.Budget = *input.Budget
	}
	if input.Duration != nil {
		existing.Duration = *input.Duration
	}
	if input.Description != nil {
		existing.Description = *input.Description
	}

	return u.repo.Update(ctx, existing)
}
// DeleteTrip checks ownership before deleting
func (u *TripUsecase) DeleteTrip(ctx context.Context, id uint, userID uint) error {
	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Security Check: Ownership verification
	if existing.UserId != userID {
		return errors.New("you do not have permission to delete this trip")
	}

	return u.repo.Delete(ctx, id)
}