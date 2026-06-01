package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"errors"
	"fmt"
	"strings"
)

type ComplaintUsecase struct {
	complaintRepo repository.ComplaintRepository
	bookingRepo   repository.BookingRepository
}

func NewComplaintUsecase(
	complaintRepo repository.ComplaintRepository,
	bookingRepo repository.BookingRepository,
) *ComplaintUsecase {
	return &ComplaintUsecase{
		complaintRepo: complaintRepo,
		bookingRepo:   bookingRepo,
	}
}

func normalizeComplaintStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "in_progress":
		return "in_progress"
	case "resolved":
		return "resolved"
	default:
		return "pending"
	}
}

func (u *ComplaintUsecase) CreateComplaint(ctx context.Context, userID uint, bookingID uint, title string, description string) (*entity.Complaint, error) {
	if userID == 0 {
		return nil, errors.New("user id is required")
	}
	if bookingID == 0 {
		return nil, errors.New("booking id is required")
	}
	if strings.TrimSpace(title) == "" || strings.TrimSpace(description) == "" {
		return nil, errors.New("title and description are required")
	}

	booking, err := u.bookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}
	if booking.UserID != userID {
		return nil, errors.New("access denied")
	}

	complaint := &entity.Complaint{
		UserID:      userID,
		BookingID:   bookingID,
		Title:       strings.TrimSpace(title),
		Description: strings.TrimSpace(description),
		Status:      "pending",
	}

	if err := u.complaintRepo.Create(ctx, complaint); err != nil {
		return nil, err
	}

	return complaint, nil
}

func (u *ComplaintUsecase) ListUserComplaints(ctx context.Context, userID uint) ([]entity.Complaint, error) {
	if userID == 0 {
		return nil, errors.New("user id is required")
	}
	return u.complaintRepo.GetByUserID(ctx, userID)
}

func (u *ComplaintUsecase) ListAllComplaints(ctx context.Context) ([]entity.Complaint, error) {
	return u.complaintRepo.GetAll(ctx)
}

func (u *ComplaintUsecase) GetComplaintByID(ctx context.Context, id uint) (*entity.Complaint, error) {
	if id == 0 {
		return nil, errors.New("complaint id is required")
	}
	return u.complaintRepo.GetByID(ctx, id)
}

func (u *ComplaintUsecase) GetComplaintByIDForUser(ctx context.Context, userID uint, role string, id uint) (*entity.Complaint, error) {
	complaint, err := u.GetComplaintByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if complaint == nil {
		return nil, fmt.Errorf("complaint not found")
	}

	if NormalizeRole(role) == "admin" {
		return complaint, nil
	}
	if complaint.UserID != userID {
		return nil, errors.New("access denied")
	}

	return complaint, nil
}

func (u *ComplaintUsecase) UpdateComplaintStatus(ctx context.Context, id uint, status string) error {
	if id == 0 {
		return errors.New("complaint id is required")
	}

	complaint, err := u.complaintRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if complaint == nil {
		return fmt.Errorf("complaint not found")
	}

	complaint.Status = normalizeComplaintStatus(status)
	return u.complaintRepo.Update(ctx, complaint)
}

func (u *ComplaintUsecase) DeleteComplaint(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("complaint id is required")
	}
	return u.complaintRepo.Delete(ctx, id)
}
