package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"errors"
	"strings"
)

type SupportUsecase struct {
	repo repository.SupportRepository
}

func NewSupportUsecase(repo repository.SupportRepository) *SupportUsecase {
	return &SupportUsecase{repo: repo}
}

func (u *SupportUsecase) CreateRequest(ctx context.Context, userID uint, subject, description string) (*entity.SupportRequest, error) {
	if strings.TrimSpace(subject) == "" {
		return nil, errors.New("subject is required")
	}
	req := &entity.SupportRequest{
		UserID:      userID,
		Subject:     subject,
		Description: description,
		Status:      "open",
	}
	if err := u.repo.CreateRequest(ctx, req); err != nil {
		return nil, err
	}
	return req, nil
}

func (u *SupportUsecase) AssignAgent(ctx context.Context, reqID uint, agentID uint, callerRole string) error {
	if NormalizeRole(callerRole) != "admin" {
		return errors.New("access denied")
	}
	req, err := u.repo.GetRequestByID(ctx, reqID)
	if err != nil {
		return err
	}
	req.AgentID = &agentID
	req.Status = "in_progress"
	return u.repo.UpdateRequest(ctx, req)
}

func (u *SupportUsecase) UpdateStatus(ctx context.Context, reqID uint, status, callerRole string) error {
	req, err := u.repo.GetRequestByID(ctx, reqID)
	if err != nil {
		return err
	}
	// Admin or assigned agent can update status
	if NormalizeRole(callerRole) != "admin" && NormalizeRole(callerRole) != "supportagent" {
		return errors.New("access denied")
	}
	req.Status = status
	return u.repo.UpdateRequest(ctx, req)
}

func (u *SupportUsecase) ListMyRequests(ctx context.Context, userID uint) ([]entity.SupportRequest, error) {
	return u.repo.ListRequests(ctx, &userID, "")
}

func (u *SupportUsecase) ListAllRequests(ctx context.Context, status, callerRole string) ([]entity.SupportRequest, error) {
	if NormalizeRole(callerRole) != "admin" && NormalizeRole(callerRole) != "supportagent" {
		return nil, errors.New("access denied")
	}
	return u.repo.ListRequests(ctx, nil, status)
}
