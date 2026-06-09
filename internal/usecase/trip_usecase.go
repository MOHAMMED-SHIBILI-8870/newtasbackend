package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"errors"
	"time"
)

type TripUsecase struct {
	repo repository.TripRepository
}

func NewTripUsecase(r repository.TripRepository) *TripUsecase {
	return &TripUsecase{repo: r}
}

func (u *TripUsecase) CreateTrip(ctx context.Context, trip *entity.Trip) error {
	if trip.From == "" || trip.To == "" {
		return errors.New("origin and destination locations are required")
	}

	if trip.Duration <= 0 {
		return errors.New("trip duration must span at least 1 day")
	}

	if trip.Price <= 0 {
		return errors.New("price is required and must be greater than 0")
	}

	if trip.StartDate.IsZero() {
		trip.StartDate = time.Now()
	}

	if trip.EndDate.IsZero() {
		trip.EndDate = trip.StartDate.AddDate(0, 0, trip.Duration-1)
	}

	return u.repo.Create(ctx, trip)
}
func (u *TripUsecase) GetTripByName(ctx context.Context, name string) (*entity.Trip, error) {
	if name == "" {
		return nil, errors.New("search trip string query cannot be empty")
	}
	return u.repo.GetByName(ctx, name)
}

func (u *TripUsecase) GetAllTrips(ctx context.Context) ([]entity.Trip, error) {
	return u.repo.GetAll(ctx)
}

func (u *TripUsecase) UpdateTrip(ctx context.Context, id uint, input entity.UpdateTripInput) error {
	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Update primitives via pointers
	if input.From != nil {
		existing.From = *input.From
	}
	if input.To != nil {
		existing.To = *input.To
	}
	if input.StartDate != nil {
		existing.StartDate = *input.StartDate
	}
	if input.EndDate != nil {
		existing.EndDate = *input.EndDate
	}
	if input.TripType != nil {
		existing.TripType = *input.TripType
	}
	if input.BudgetLevel != nil {
		existing.BudgetLevel = *input.BudgetLevel
	}
	if input.HotelType != nil {
		existing.HotelType = *input.HotelType
	}
	if input.Transport != nil {
		existing.Transport = *input.Transport
	}
	if input.ItineraryRaw != nil {
		existing.ItineraryRaw = *input.ItineraryRaw
	}
	if input.ImageURL != nil {
		existing.ImageURL = *input.ImageURL
	}
	if input.Status != nil {
		existing.Status = *input.Status
	}

	if input.Duration != nil {
		if *input.Duration <= 0 {
			return errors.New("duration must be at least 1 day")
		}
		existing.Duration = *input.Duration
	}
	if input.Price != nil {
		if *input.Price < 0 {
			return errors.New("price cannot be negative")
		}
		existing.Price = *input.Price
	}
	if input.Members != nil {
		existing.Members = *input.Members
	}
	if input.Children != nil {
		existing.Children = *input.Children
	}

	// Override specific timeline structures directly if updated
	if input.Plans != nil {
		existing.Plans = input.Plans
	}

	return u.repo.Update(ctx, existing)
}

func (u *TripUsecase) DeleteTrip(ctx context.Context, id uint) error {
	_, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return u.repo.Delete(ctx, id)
}
