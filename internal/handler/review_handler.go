package handler

import (
	"backend/internal/dto"
	"backend/internal/entity"
	"backend/internal/middleware"
	"backend/internal/response"
	"backend/internal/usecase"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ReviewHandler struct {
	usecase *usecase.ReviewUsecase
}

func NewReviewHandler(usecase *usecase.ReviewUsecase) *ReviewHandler {
	return &ReviewHandler{usecase: usecase}
}

func reviewStatusFromErr(err error) int {
	if err == nil {
		return fiber.StatusBadRequest
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "not found"):
		return fiber.StatusNotFound
	case strings.Contains(msg, "access denied"):
		return fiber.StatusForbidden
	case strings.Contains(msg, "already exists"):
		return fiber.StatusConflict
	default:
		return fiber.StatusBadRequest
	}
}

func mapReviewResponse(review entity.Review) dto.ReviewResponse {
	resp := dto.ReviewResponse{
		ID:        review.ID,
		UserID:    review.UserID,
		TripID:    review.TripID,
		GuideID:   review.GuideID,
		Rating:    review.Rating,
		Comment:   review.Comment,
		CreatedAt: review.CreatedAt,
		UpdatedAt: review.UpdatedAt,
	}

	if review.User.ID != 0 {
		resp.User = &dto.ReviewUserResponse{
			FullName: review.User.FullName,
			Email:    review.User.Email,
		}
	}

	if review.Trip.ID != 0 {
		resp.Trip = &dto.ReviewTripResponse{
			From: review.Trip.From,
			To:   review.Trip.To,
		}
	}

	return resp
}

func (h *ReviewHandler) CreateReview(c *fiber.Ctx) error {
	var input dto.ReviewRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	userID := middleware.GetAuthUserID(c)
	review, err := h.usecase.CreateReview(c.Context(), userID, input.TripID, input.GuideID, input.Rating, input.Comment)
	if err != nil {
		return response.Fail(c, reviewStatusFromErr(err), "failed to create review", err)
	}

	return response.Success(c, fiber.StatusCreated, "review created successfully", mapReviewResponse(*review))
}

func (h *ReviewHandler) ListTripReviews(c *fiber.Ctx) error {
	tripID, err := strconv.Atoi(c.Params("trip_id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid trip id", err)
	}

	reviews, err := h.usecase.ListTripReviews(c.Context(), uint(tripID))
	if err != nil {
		return response.Fail(c, reviewStatusFromErr(err), "failed to load reviews", err)
	}

	items := make([]dto.ReviewResponse, 0, len(reviews))
	for _, review := range reviews {
		items = append(items, mapReviewResponse(review))
	}

	return response.Success(c, fiber.StatusOK, "reviews loaded successfully", items)
}

func (h *ReviewHandler) ListMyReviews(c *fiber.Ctx) error {
	userID := middleware.GetAuthUserID(c)
	reviews, err := h.usecase.ListUserReviews(c.Context(), userID)
	if err != nil {
		return response.Fail(c, reviewStatusFromErr(err), "failed to load reviews", err)
	}

	items := make([]dto.ReviewResponse, 0, len(reviews))
	for _, review := range reviews {
		items = append(items, mapReviewResponse(review))
	}

	return response.Success(c, fiber.StatusOK, "user reviews loaded successfully", items)
}

func (h *ReviewHandler) ListAllReviews(c *fiber.Ctx) error {
	reviews, err := h.usecase.ListAllReviews(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to load reviews", err)
	}

	items := make([]dto.ReviewResponse, 0, len(reviews))
	for _, review := range reviews {
		items = append(items, mapReviewResponse(review))
	}

	return response.Success(c, fiber.StatusOK, "reviews loaded successfully", items)
}

func (h *ReviewHandler) GetTripSummary(c *fiber.Ctx) error {
	tripID, err := strconv.Atoi(c.Params("trip_id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid trip id", err)
	}

	average, count, err := h.usecase.GetTripAverageRating(c.Context(), uint(tripID))
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to load review summary", err)
	}

	return response.Success(c, fiber.StatusOK, "review summary loaded successfully", dto.ReviewSummaryResponse{
		TripID:        uint(tripID),
		AverageRating: average,
		ReviewCount:   count,
	})
}

func (h *ReviewHandler) DeleteReview(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid review id", err)
	}

	if err := h.usecase.DeleteReview(c.Context(), uint(id)); err != nil {
		return response.Fail(c, reviewStatusFromErr(err), "failed to delete review", err)
	}

	return response.Success(c, fiber.StatusOK, "review deleted successfully", nil)
}
