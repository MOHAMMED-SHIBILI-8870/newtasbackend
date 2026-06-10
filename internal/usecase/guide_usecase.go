package usecase

import (
	"backend/internal/dto"
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
)

type GuideUsecase struct {
	repo repository.GuideRepository
}

func NewGuideUsecase(repo repository.GuideRepository) *GuideUsecase {
	return &GuideUsecase{
		repo: repo,
	}
}

func (u *GuideUsecase) GetProfile(ctx context.Context,userID uint) (*entity.Guide, error) {

	return u.repo.GetProfile(
		ctx,
		userID,
	)
}

func (u *GuideUsecase) UpdateProfile(ctx context.Context,userID uint,input dto.UpdateGuideProfileInput) error {

	guide, err := u.repo.GetProfile(
		ctx,
		userID,
	)

	if err != nil {
		return err
	}

	guide.Bio = input.Bio
	guide.Experience = input.Experience
	guide.Languages = input.Languages
	guide.IsAvailable = input.IsAvailable

	return u.repo.UpdateProfile(
		ctx,
		guide,
	)
}

