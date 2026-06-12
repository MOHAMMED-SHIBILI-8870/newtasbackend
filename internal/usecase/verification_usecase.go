package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
)

type VerificationUsecase interface {
	SubmitVerification(verification *entity.Verification) error
	GetVerificationByBookingID(bookingID uint) (*entity.Verification, error)
}

type verificationUsecase struct {
	verificationRepo repository.VerificationRepository
}

func NewVerificationUsecase(verificationRepo repository.VerificationRepository) VerificationUsecase {
	return &verificationUsecase{
		verificationRepo: verificationRepo,
	}
}

func (u *verificationUsecase) SubmitVerification(verification *entity.Verification) error {
	// Any validation logic could go here
	return u.verificationRepo.Create(verification)
}

func (u *verificationUsecase) GetVerificationByBookingID(bookingID uint) (*entity.Verification, error) {
	return u.verificationRepo.GetByBookingID(bookingID)
}
