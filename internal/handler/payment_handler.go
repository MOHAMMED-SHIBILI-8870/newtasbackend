package handler

import (
	"backend/internal/response"
	"backend/internal/usecase"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type PaymentHandler struct {
	usecase *usecase.PaymentUsecase
}

func NewPaymentHandler(u *usecase.PaymentUsecase) *PaymentHandler {
	return &PaymentHandler{
		usecase: u,
	}
}

type VerifyPaymentInput struct {
	OrderID   string `json:"order_id"`
	PaymentID string `json:"payment_id"`
	Signature string `json:"signature"`
}

func (h *PaymentHandler) CreateAdvancePayment(c *fiber.Ctx) error {
	bookingID, err := strconv.Atoi(c.Params("booking_id"))
	fmt.Println("BOOKING ID FROM URL:", bookingID)

	if err != nil {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			"invalid booking id",
			err,
		)
	}

	paymentData, err := h.usecase.CreateAdvancePayment(c.Context(), uint(bookingID))
	if err != nil {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			"failed to create advance payment",
			err,
		)
	}

	return response.Success(c, fiber.StatusCreated, "advance payment created successfully", paymentData)
}

func (h *PaymentHandler) VerifyAdvancePayment(c *fiber.Ctx) error {
	bookingID, err := strconv.Atoi(c.Params("booking_id"))
	if err != nil {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			"invalid booking id",
			err,
		)
	}

	var input VerifyPaymentInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			"invalid request body",
			err,
		)
	}

	err = h.usecase.VerifyAdvancePayment(
		c.Context(),
		uint(bookingID),
		input.OrderID,
		input.PaymentID,
		input.Signature,
	)
	if err != nil {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			"payment verification failed",
			err,
		)
	}

	return response.Success(
		c,
		fiber.StatusOK,
		"advance payment verified successfully",
		nil,
	)
}

func (h *PaymentHandler) CreateBalancePayment(c *fiber.Ctx) error {
	bookingID, err := strconv.Atoi(c.Params("booking_id"))
	if err != nil {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			"invalid booking id",
			err,
		)
	}

	payment, err := h.usecase.CreateBalancePayment(
		c.Context(),
		uint(bookingID),
	)
	if err != nil {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			"failed to create balance payment",
			err,
		)
	}

	return response.Success(
		c,
		fiber.StatusCreated,
		"balance payment created successfully",
		payment,
	)
}

func (h *PaymentHandler) VerifyBalancePayment(c *fiber.Ctx) error {
	bookingID, err := strconv.Atoi(c.Params("booking_id"))
	if err != nil {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			"invalid booking id",
			err,
		)
	}

	var input VerifyPaymentInput
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			"invalid request body",
			err,
		)
	}

	err = h.usecase.VerifyBalancePayment(
		c.Context(),
		uint(bookingID),
		input.OrderID,
		input.PaymentID,
		input.Signature,
	)
	if err != nil {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			"balance payment verification failed",
			err,
		)
	}

	return response.Success(
		c,
		fiber.StatusOK,
		"balance payment verified successfully",
		nil,
	)
}