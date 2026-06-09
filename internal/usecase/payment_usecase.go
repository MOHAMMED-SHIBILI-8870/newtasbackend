package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"backend/internal/service"
	"context"
	"errors"
	"fmt"
	"os"
)

type PaymentUsecase struct {
	paymentRepo repository.PaymentRepository
	bookingRepo repository.BookingRepository
	razorpay    *service.RazorpayService
}

func NewPaymentUsecase(
	paymentRepo repository.PaymentRepository,
	bookingRepo repository.BookingRepository,
	razorpay *service.RazorpayService,
) *PaymentUsecase {
	return &PaymentUsecase{
		paymentRepo: paymentRepo,
		bookingRepo: bookingRepo,
		razorpay:    razorpay,
	}
}

func (u *PaymentUsecase) CreateAdvancePayment(ctx context.Context, bookingID uint) (map[string]interface{}, error) {
	booking, err := u.bookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	if booking.PaymentStatus == "fully_paid" {
		return nil, errors.New("booking already fully paid")
	}

	order, err := u.razorpay.CreateOrder(booking.AdvanceAmount, fmt.Sprintf("booking_%d_advance", booking.ID))
	if err != nil {
		return nil, err
	}

	payment := &entity.Payment{
		BookingID:       booking.ID,
		PaymentType:     "advance",
		Status:          "pending",
		Amount:          booking.AdvanceAmount,
		RazorpayOrderID: order["id"].(string),
	}

	if err := u.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}

	responsePayload := map[string]interface{}{
		"payment":         payment,
		"razorpay_key_id": os.Getenv("RAZORPAY_KEY_ID"),
	}

	return responsePayload, nil
}

func (u *PaymentUsecase) VerifyAdvancePayment(ctx context.Context, bookingID uint, orderID, paymentID, signature string) error {
	if !u.razorpay.VerifySignature(orderID, paymentID, signature) {
		return errors.New("invalid cryptographic razorpay signature hash")
	}

	booking, err := u.bookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return err
	}

	payment, err := u.paymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return err
	}

	payment.Status = "success"
	payment.RazorpayPaymentID = paymentID

	if err := u.paymentRepo.Update(ctx, payment); err != nil {
		return err
	}

	booking.Status = "confirmed"
	booking.PaymentStatus = "advance_paid"

	return u.bookingRepo.UpdateBooking(ctx, booking)
}

func (u *PaymentUsecase) CreateBalancePayment(ctx context.Context, bookingID uint) (map[string]interface{}, error) {
	booking, err := u.bookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	// 🔒 Guards tracking rules logic sequence
	if booking.PaymentStatus != "advance_paid" {
		return nil, errors.New("advance token checkpoint payment required first")
	}

	if booking.BalanceAmount <= 0 {
		return nil, errors.New("no balance remaining on checkout records")
	}

	order, err := u.razorpay.CreateOrder(booking.BalanceAmount, fmt.Sprintf("booking_%d_balance", booking.ID))
	if err != nil {
		return nil, err
	}

	payment := &entity.Payment{
		BookingID:       booking.ID,
		PaymentType:     "balance",
		Status:          "pending",
		Amount:          booking.BalanceAmount,
		RazorpayOrderID: order["id"].(string),
	}

	if err := u.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}

	responsePayload := map[string]interface{}{
		"payment":         payment,
		"razorpay_key_id": os.Getenv("RAZORPAY_KEY_ID"),
	}

	return responsePayload, nil
}

func (u *PaymentUsecase) VerifyBalancePayment(ctx context.Context, bookingID uint, orderID, paymentID, signature string) error {
	if !u.razorpay.VerifySignature(orderID, paymentID, signature) {
		return errors.New("invalid cryptographic razorpay signature hash")
	}

	booking, err := u.bookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return err
	}

	payment, err := u.paymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return err
	}

	payment.Status = "success"
	payment.RazorpayPaymentID = paymentID

	if err := u.paymentRepo.Update(ctx, payment); err != nil {
		return err
	}

	booking.PaymentStatus = "fully_paid"

	return u.bookingRepo.UpdateBooking(ctx, booking)
}