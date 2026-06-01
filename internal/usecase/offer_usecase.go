package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type OfferUsecase struct {
	repo repository.OfferRepository
}

func NewOfferUsecase(repo repository.OfferRepository) *OfferUsecase {
	return &OfferUsecase{repo: repo}
}

func normalizeOfferCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func (u *OfferUsecase) CreateOffer(ctx context.Context, offer *entity.Offer) error {
	if offer == nil {
		return errors.New("offer is required")
	}

	offer.Code = normalizeOfferCode(offer.Code)
	offer.Title = strings.TrimSpace(offer.Title)
	if offer.Code == "" || offer.Title == "" {
		return errors.New("offer code and title are required")
	}
	if offer.DiscountPercent <= 0 || offer.DiscountPercent > 100 {
		return errors.New("discount percent must be between 0 and 100")
	}
	if offer.ExpiryDate.IsZero() || offer.ExpiryDate.Before(time.Now()) {
		return errors.New("expiry date must be in the future")
	}

	existing, err := u.repo.GetByCode(ctx, offer.Code)
	if err != nil {
		return err
	}
	if existing != nil {
		return fmt.Errorf("offer already exists")
	}

	if strings.TrimSpace(offer.Description) == "" {
		offer.Description = offer.Title
	}

	return u.repo.Create(ctx, offer)
}

func (u *OfferUsecase) UpdateOffer(ctx context.Context, id uint, offer *entity.Offer) error {
	if id == 0 {
		return errors.New("offer id is required")
	}
	if offer == nil {
		return errors.New("offer is required")
	}

	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("offer not found")
	}

	if strings.TrimSpace(offer.Code) != "" {
		existing.Code = normalizeOfferCode(offer.Code)
	}
	if strings.TrimSpace(offer.Title) != "" {
		existing.Title = strings.TrimSpace(offer.Title)
	}
	if strings.TrimSpace(offer.Description) != "" {
		existing.Description = strings.TrimSpace(offer.Description)
	}
	if offer.DiscountPercent > 0 {
		existing.DiscountPercent = offer.DiscountPercent
	}
	if !offer.ExpiryDate.IsZero() {
		existing.ExpiryDate = offer.ExpiryDate
	}
	existing.Active = offer.Active

	if existing.DiscountPercent <= 0 || existing.DiscountPercent > 100 {
		return errors.New("discount percent must be between 0 and 100")
	}
	if existing.ExpiryDate.Before(time.Now()) {
		return errors.New("expiry date must be in the future")
	}

	return u.repo.Update(ctx, existing)
}

func (u *OfferUsecase) DeleteOffer(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("offer id is required")
	}
	return u.repo.Delete(ctx, id)
}

func (u *OfferUsecase) ListOffers(ctx context.Context) ([]entity.Offer, error) {
	return u.repo.List(ctx)
}

func (u *OfferUsecase) GetActiveOffers(ctx context.Context) ([]entity.Offer, error) {
	return u.repo.GetActive(ctx)
}

func (u *OfferUsecase) GetOfferByCode(ctx context.Context, code string) (*entity.Offer, error) {
	normalized := normalizeOfferCode(code)
	if normalized == "" {
		return nil, errors.New("offer code is required")
	}
	return u.repo.GetByCode(ctx, normalized)
}

func (u *OfferUsecase) ValidateCoupon(ctx context.Context, code string) (*entity.Offer, error) {
	offer, err := u.GetOfferByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if offer == nil {
		return nil, fmt.Errorf("offer not found")
	}
	if !offer.Active {
		return nil, errors.New("offer is inactive")
	}
	if offer.ExpiryDate.Before(time.Now()) {
		return nil, errors.New("offer has expired")
	}
	return offer, nil
}

func (u *OfferUsecase) ApplyDiscount(amount float64, offer *entity.Offer) float64 {
	if amount <= 0 || offer == nil {
		return amount
	}
	discount := amount * (offer.DiscountPercent / 100)
	finalAmount := amount - discount
	if finalAmount < 0 {
		return 0
	}
	return math.Round(finalAmount*100) / 100
}
