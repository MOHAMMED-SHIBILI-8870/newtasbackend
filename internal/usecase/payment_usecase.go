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

func (u *PaymentUsecase) CreateAdvancePayment(ctx context.Context, bookingID uint, callerID uint) (map[string]interface{}, error) {
	booking, err := u.bookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	if booking.UserID != callerID {
		return nil, errors.New("access denied")
	}

	if booking.PaymentStatus == "fully_paid" {
		return nil, errors.New("booking already fully paid")
	}

	if booking.PaymentStatus == "advance_paid" {
		return nil, errors.New("advance payment already completed")
	}

	// Prevent duplicate pending advance
	existingPayment, err := u.paymentRepo.GetPendingPayment(ctx,booking.ID,"advance")

	if err == nil && existingPayment != nil {
		return nil, errors.New("advance payment is already pending")
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

func (u *PaymentUsecase) VerifyAdvancePayment(ctx context.Context, bookingID uint, callerID uint, orderID, paymentID, signature string) error {
	if !u.razorpay.VerifySignature(orderID, paymentID, signature) {
		return errors.New("invalid cryptographic razorpay signature hash")
	}

	booking, err := u.bookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.UserID != callerID {
		return errors.New("access denied")
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

func (u *PaymentUsecase) CreateBalancePayment(ctx context.Context, bookingID uint, callerID uint) (map[string]interface{}, error) {
	booking, err := u.bookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	if booking.UserID != callerID {
		return nil, errors.New("access denied")
	}

	// 🔒 Guards tracking rules logic sequence
	if booking.PaymentStatus != "advance_paid" {
		return nil, errors.New("advance token checkpoint payment required first")
	}

	// Prevent duplicate pending balance
	existingPayment, err := u.paymentRepo.GetPendingPayment(ctx,booking.ID,"balance")

if err == nil && existingPayment != nil {
	return nil, errors.New("balance payment is already pending")
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

func (u *PaymentUsecase) VerifyBalancePayment(ctx context.Context, bookingID uint, callerID uint, orderID, paymentID, signature string) error {
	if !u.razorpay.VerifySignature(orderID, paymentID, signature) {
		return errors.New("invalid cryptographic razorpay signature hash")
	}

	booking, err := u.bookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return err
	}

	if booking.UserID != callerID {
		return errors.New("access denied")
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

func (u *PaymentUsecase) RefundPayment(ctx context.Context, paymentID uint, callerRole string) error {
	if NormalizeRole(callerRole) != "admin" {
		return errors.New("access denied")
	}

	payment, err := u.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	if payment.Status != "success" && payment.Status != "paid" {
		return errors.New("payment is not eligible for refund")
	}
	if payment.RazorpayPaymentID == "" {
		return errors.New("razorpay payment id missing")
	}

	_, err = u.razorpay.RefundPayment(payment.RazorpayPaymentID, payment.Amount)
	if err != nil {
		return err
	}

	payment.Status = "refunded"
	if err := u.paymentRepo.Update(ctx, payment); err != nil {
		return err
	}

	booking, err := u.bookingRepo.GetBookingByID(ctx, payment.BookingID)
	if err == nil {
		booking.PaymentStatus = "refunded"
		booking.Status = "cancelled"
		_ = u.bookingRepo.UpdateBooking(ctx, booking)
	}

	return nil
}

func (u *PaymentUsecase) GetPaymentHistory(ctx context.Context, callerID uint) ([]entity.Payment, error) {
	return u.paymentRepo.GetHistoryByUserID(ctx, callerID)
}

func (u *PaymentUsecase) GetInvoice(ctx context.Context, paymentID uint, callerID uint, callerRole string) (map[string]interface{}, error) {
	payment, err := u.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	if NormalizeRole(callerRole) != "admin" && payment.Booking.UserID != callerID {
		return nil, errors.New("access denied")
	}

	invoice := map[string]interface{}{
		"invoice_id":     fmt.Sprintf("INV-%d", payment.ID),
		"date":           payment.UpdatedAt,
		"booking_id":     payment.BookingID,
		"payment_type":   payment.PaymentType,
		"status":         payment.Status,
		"amount_paid":    payment.Amount,
		"transaction_id": payment.RazorpayPaymentID,
	}

	return invoice, nil
}
