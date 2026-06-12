package handler

import (
	"backend/internal/dto"
	"backend/internal/entity"
	"backend/internal/response"
	"backend/internal/usecase"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type OfferHandler struct {
	usecase *usecase.OfferUsecase
}

func NewOfferHandler(usecase *usecase.OfferUsecase) *OfferHandler {
	return &OfferHandler{usecase: usecase}
}

func offerStatusFromErr(err error) int {
	if err == nil {
		return fiber.StatusBadRequest
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "not found"):
		return fiber.StatusNotFound
	case strings.Contains(msg, "already exists"):
		return fiber.StatusConflict
	default:
		return fiber.StatusBadRequest
	}
}

func mapOfferResponse(offer entity.Offer) dto.OfferResponse {
	return dto.OfferResponse{
		ID:              offer.ID,
		Code:            offer.Code,
		Title:           offer.Title,
		Description:     offer.Description,
		DiscountPercent: offer.DiscountPercent,
		DiscountType:    offer.DiscountType,
		FixedDiscount:   offer.FixedDiscount,
		MaxUsage:        offer.MaxUsage,
		CurrentUsage:    offer.CurrentUsage,
		TripID:          offer.TripID,
		ExpiryDate:      offer.ExpiryDate,
		Active:          offer.Active,
		CreatedAt:       offer.CreatedAt,
		UpdatedAt:       offer.UpdatedAt,
	}
}

func (h *OfferHandler) ListOffers(c *fiber.Ctx) error {
	offers, err := h.usecase.ListOffers(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to load offers", err)
	}

	items := make([]dto.OfferResponse, 0, len(offers))
	for _, offer := range offers {
		items = append(items, mapOfferResponse(offer))
	}

	return response.Success(c, fiber.StatusOK, "offers loaded successfully", items)
}

func (h *OfferHandler) ListActiveOffers(c *fiber.Ctx) error {
	offers, err := h.usecase.GetActiveOffers(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to load active offers", err)
	}

	items := make([]dto.OfferResponse, 0, len(offers))
	for _, offer := range offers {
		items = append(items, mapOfferResponse(offer))
	}

	return response.Success(c, fiber.StatusOK, "active offers loaded successfully", items)
}

func (h *OfferHandler) CreateOffer(c *fiber.Ctx) error {
	var input dto.OfferRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	offer := &entity.Offer{
		Code:            input.Code,
		Title:           input.Title,
		Description:     input.Description,
		DiscountPercent: input.DiscountPercent,
		DiscountType:    input.DiscountType,
		FixedDiscount:   input.FixedDiscount,
		MaxUsage:        input.MaxUsage,
		TripID:          input.TripID,
		ExpiryDate:      input.ExpiryDate,
		Active:          input.Active,
	}

	if err := h.usecase.CreateOffer(c.Context(), offer); err != nil {
		return response.Fail(c, offerStatusFromErr(err), "failed to create offer", err)
	}

	return response.Success(c, fiber.StatusCreated, "offer created successfully", mapOfferResponse(*offer))
}

func (h *OfferHandler) UpdateOffer(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid offer id", err)
	}

	var input dto.OfferRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	offer := &entity.Offer{
		Code:            input.Code,
		Title:           input.Title,
		Description:     input.Description,
		DiscountPercent: input.DiscountPercent,
		DiscountType:    input.DiscountType,
		FixedDiscount:   input.FixedDiscount,
		MaxUsage:        input.MaxUsage,
		TripID:          input.TripID,
		ExpiryDate:      input.ExpiryDate,
		Active:          input.Active,
	}

	if err := h.usecase.UpdateOffer(c.Context(), uint(id), offer); err != nil {
		return response.Fail(c, offerStatusFromErr(err), "failed to update offer", err)
	}

	return response.Success(c, fiber.StatusOK, "offer updated successfully", nil)
}

func (h *OfferHandler) DeleteOffer(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid offer id", err)
	}

	if err := h.usecase.DeleteOffer(c.Context(), uint(id)); err != nil {
		return response.Fail(c, offerStatusFromErr(err), "failed to delete offer", err)
	}

	return response.Success(c, fiber.StatusOK, "offer deleted successfully", nil)
}

func (h *OfferHandler) ValidateCoupon(c *fiber.Ctx) error {
	var input dto.ApplyCouponRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	offer, err := h.usecase.ValidateCoupon(c.Context(), input.Code)
	if err != nil {
		return response.Fail(c, offerStatusFromErr(err), "invalid coupon", err)
	}

	return response.Success(c, fiber.StatusOK, "coupon is valid", mapOfferResponse(*offer))
}
