package handler

import (
	"backend/internal/middleware"
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

	userID := middleware.GetAuthUserID(c)
	if userID == 0 {
		return response.Fail(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	paymentData, err := h.usecase.CreateAdvancePayment(c.Context(), uint(bookingID), userID)
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

	userID := middleware.GetAuthUserID(c)
	if userID == 0 {
		return response.Fail(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	err = h.usecase.VerifyAdvancePayment(
		c.Context(),
		uint(bookingID),
		userID,
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

	userID := middleware.GetAuthUserID(c)
	if userID == 0 {
		return response.Fail(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	payment, err := h.usecase.CreateBalancePayment(
		c.Context(),
		uint(bookingID),
		userID,
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

	userID := middleware.GetAuthUserID(c)
	if userID == 0 {
		return response.Fail(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	err = h.usecase.VerifyBalancePayment(
		c.Context(),
		uint(bookingID),
		userID,
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

func (h *PaymentHandler) RefundPayment(c *fiber.Ctx) error {
	paymentID, err := strconv.Atoi(c.Params("payment_id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid payment id", err)
	}

	role := middleware.GetAuthRole(c)
	err = h.usecase.RefundPayment(c.Context(), uint(paymentID), role)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "failed to refund payment", err)
	}

	return response.Success(c, fiber.StatusOK, "payment refunded successfully", nil)
}

func (h *PaymentHandler) GetPaymentHistory(c *fiber.Ctx) error {
	userID := middleware.GetAuthUserID(c)
	if userID == 0 {
		return response.Fail(c, fiber.StatusUnauthorized, "unauthorized", nil)
	}

	history, err := h.usecase.GetPaymentHistory(c.Context(), userID)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to load payment history", err)
	}

	return response.Success(c, fiber.StatusOK, "payment history loaded successfully", history)
}

func (h *PaymentHandler) GetInvoice(c *fiber.Ctx) error {
	paymentID, err := strconv.Atoi(c.Params("payment_id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid payment id", err)
	}

	userID := middleware.GetAuthUserID(c)
	role := middleware.GetAuthRole(c)

	invoice, err := h.usecase.GetInvoice(c.Context(), uint(paymentID), userID, role)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "failed to get invoice", err)
	}

	return response.Success(c, fiber.StatusOK, "invoice generated successfully", invoice)
}