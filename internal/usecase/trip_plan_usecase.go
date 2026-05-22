package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"errors"
)

type TripPlanUsecase struct {
	repo repository.TripPlanRepository
}

func NewTripPlanUsecase(r repository.TripPlanRepository) *TripPlanUsecase {
	return &TripPlanUsecase{repo: r}
}

// CREATE PLAN
func (u *TripPlanUsecase) CreateTripPlan(ctx context.Context, plan *entity.TripPlan) error {

	if plan.TripID == 0 {
		return errors.New("trip_id is required")
	}

	if plan.DayNumber <= 0 {
		return errors.New("day_number must be greater than 0")
	}

	if plan.Title == "" {
		return errors.New("title is required")
	}

	return u.repo.Create(ctx, plan)
}

// GET FULL ITINERARY
func (u *TripPlanUsecase) GetTripPlans(ctx context.Context, tripID uint) ([]entity.TripPlan, error) {

	if tripID == 0 {
		return nil, errors.New("trip_id is required")
	}

	return u.repo.GetByTripID(ctx, tripID)
}

// DELETE PLAN
func (u *TripPlanUsecase) DeleteTripPlan(ctx context.Context, id uint) error {

	if id == 0 {
		return errors.New("invalid id")
	}

	return u.repo.Delete(ctx, id)
}
