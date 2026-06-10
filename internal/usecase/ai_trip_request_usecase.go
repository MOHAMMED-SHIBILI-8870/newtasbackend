package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AITripRequestUsecase struct {
	repo                repository.AITripRequestRepository
	tripRepo            repository.TripRepository
	userRepo            repository.UserRepository
	notificationUsecase *NotificationUsecase
	db                  *gorm.DB
}

func NewAITripRequestUsecase(
	repo repository.AITripRequestRepository,
	tripRepo repository.TripRepository,
	userRepo repository.UserRepository,
	notificationUsecase *NotificationUsecase,
	db *gorm.DB,
) *AITripRequestUsecase {
	return &AITripRequestUsecase{
		repo:                repo,
		tripRepo:            tripRepo,
		userRepo:            userRepo,
		notificationUsecase: notificationUsecase,
		db:                  db,
	}
}

func (u *AITripRequestUsecase) CreateRequest(ctx context.Context, userID uint, input entity.AITripRequestInput) (*entity.AITripRequest, error) {
	if userID == 0 {
		return nil, errors.New("user is required")
	}
	if strings.TrimSpace(input.From) == "" || strings.TrimSpace(input.To) == "" {
		return nil, errors.New("from and to locations are required")
	}
	if input.Days <= 0 {
		input.Days = 1
	}
	if input.Members <= 0 {
		input.Members = 1
	}
	if input.Children < 0 {
		input.Children = 0
	}
	if strings.TrimSpace(input.GeneratedPlan) == "" {
		return nil, errors.New("generated plan is required")
	}
	if strings.TrimSpace(input.Prompt) == "" {
		input.Prompt = fmt.Sprintf("AI trip request from %s to %s", input.From, input.To)
	}

	request := &entity.AITripRequest{
		UserID:        userID,
		From:          strings.TrimSpace(input.From),
		To:            strings.TrimSpace(input.To),
		Days:          input.Days,
		TripType:      normalizeTripType(input.TripType),
		BudgetLevel:   normalizeTripBudget(input.BudgetLevel),
		Members:       input.Members,
		Children:      input.Children,
		HotelType:     normalizeTripHotel(input.HotelType),
		Transport:     normalizeTripTransport(input.Transport),
		Prompt:        input.Prompt,
		GeneratedPlan: input.GeneratedPlan,
		Status:        entity.AITripStatusPending,
	}

	if err := u.repo.Create(ctx, request); err != nil {
		return nil, err
	}

	if user, err := u.userRepo.GetByID(ctx, userID); err == nil && user != nil {
		request.User = *user
	}

	if u.notificationUsecase != nil && u.userRepo != nil {
		admins, err := u.userRepo.GetUsers(ctx, 0,"admin", "")
		if err == nil {
			for _, admin := range admins {
				title := "AI trip request pending review"
				message := fmt.Sprintf("%s submitted a trip request from %s to %s.",
					strings.TrimSpace(input.Prompt), request.From, request.To)
				if err := u.notificationUsecase.CreateNotification(ctx, &entity.Notification{
					UserID:          admin.ID,
					Type:            "ai_request",
					Title:           title,
					Message:         message,
					AITripRequestID: &request.ID,
				}); err != nil {
					log.Printf("failed to notify admin %d about ai request %d: %v", admin.ID, request.ID, err)
				}
			}
		}
	}

	return request, nil
}

func (u *AITripRequestUsecase) GetRequests(ctx context.Context, role string, userID uint) ([]entity.AITripRequest, error) {
	if NormalizeRole(role) == "admin" {
		return u.repo.GetAll(ctx)
	}
	if userID == 0 {
		return nil, errors.New("user is required")
	}
	return u.repo.GetByUserID(ctx, userID)
}

func (u *AITripRequestUsecase) ReviewRequest(ctx context.Context, adminID, requestID uint, approve bool, adminNote string) (*entity.AITripRequest, error) {
	if adminID == 0 {
		return nil, errors.New("admin is required")
	}
	if requestID == 0 {
		return nil, errors.New("request id is required")
	}
	if u.db == nil {
		return nil, errors.New("database unavailable")
	}

	var reviewedID uint
	if err := u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var request entity.AITripRequest
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("User").
			Preload("Trip").
			Preload("ReviewedBy").
			First(&request, requestID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("ai trip request not found")
			}
			return err
		}
		if request.Status != entity.AITripStatusPending {
			return errors.New("request has already been reviewed")
		}

		now := time.Now().UTC()
		request.ReviewedByID = &adminID
		request.ReviewedAt = &now
		request.AdminNote = adminNote

		if approve {
			trip := &entity.Trip{
				From:         request.From,
				To:           request.To,
				StartDate:    now,
				EndDate:      now.AddDate(0, 0, maxInt(request.Days, 1)),
				Duration:     maxInt(request.Days, 1),
				TripType:     request.TripType,
				BudgetLevel:  request.BudgetLevel,
				Price:        0,
				Members:      request.Members,
				Children:     request.Children,
				HotelType:    request.HotelType,
				Transport:    request.Transport,
				ItineraryRaw: request.GeneratedPlan,
				Status:       "active",
			}

			if err := tx.Create(trip).Error; err != nil {
				return err
			}

			request.TripID = &trip.ID
			request.Status = entity.AITripStatusApproved
		} else {
			request.Status = entity.AITripStatusRejected
		}

		if err := tx.Save(&request).Error; err != nil {
			return err
		}

		if u.notificationUsecase != nil {
			title := "AI trip request updated"
			message := "Your AI trip request has been reviewed."
			if approve {
				title = "AI trip approved"
				message = fmt.Sprintf("Your AI trip request from %s to %s has been approved.", request.From, request.To)
			} else {
				title = "AI trip rejected"
				message = fmt.Sprintf("Your AI trip request from %s to %s has been rejected.", request.From, request.To)
			}

			if err := tx.Create(&entity.Notification{
				UserID:          request.UserID,
				Type:            "ai_request",
				Title:           title,
				Message:         message,
				AITripRequestID: &request.ID,
				IsRead:          false,
			}).Error; err != nil {
				return err
			}
		}

		reviewedID = request.ID
		return nil
	}); err != nil {
		return nil, err
	}

	request, err := u.repo.GetByID(ctx, reviewedID)
	if err != nil {
		return nil, err
	}
	if request == nil {
		return nil, errors.New("ai trip request not found")
	}

	return request, nil
}

func normalizeTripType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "adventure":
		return "Adventure"
	case "solo":
		return "Solo"
	case "couple":
		return "Couple"
	case "friends":
		return "Friends"
	default:
		return "Family"
	}
}

func normalizeTripBudget(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "low":
		return "Low"
	case "high":
		return "High"
	default:
		return "Medium"
	}
}

func normalizeTripHotel(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "budget":
		return "Budget"
	case "4 star":
		return "4 Star"
	case "5 star":
		return "5 Star"
	default:
		return "3 Star"
	}
}

func normalizeTripTransport(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "bus":
		return "Bus"
	case "train":
		return "Train"
	case "flight":
		return "Flight"
	default:
		return "Car"
	}
}

func maxInt(values ...int) int {
	max := 0
	for _, value := range values {
		if value > max {
			max = value
		}
	}
	if max == 0 {
		return 1
	}
	return max
}
