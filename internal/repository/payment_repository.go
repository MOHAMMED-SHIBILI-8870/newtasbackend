package repository

import (
	"backend/internal/entity"
	"context"

	"gorm.io/gorm"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *entity.Payment)error
	Update(ctx context.Context, payment *entity.Payment)error
	GetByOrderID(ctx context.Context, orderID string)(*entity.Payment,error)
	GetByID(ctx context.Context, id uint) (*entity.Payment, error)
	GetHistoryByUserID(ctx context.Context, userID uint) ([]entity.Payment, error)
	GetPendingPayment(ctx context.Context,bookingID uint,paymentType string) (*entity.Payment, error)
}



type paymentRepository struct{
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB)PaymentRepository{
	return &paymentRepository{db : db}
}


func (r *paymentRepository) Create(ctx context.Context,payment *entity.Payment)error{
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *paymentRepository)Update(ctx context.Context,payment *entity.Payment)error{
	return r.db.WithContext(ctx).Save(payment).Error
}

func (r *paymentRepository)GetByOrderID(ctx context.Context,orderID string)(*entity.Payment,error){
	var payment entity.Payment

	err:= r.db.WithContext(ctx).Where("razorpay_order_id=?",orderID).First(&payment).Error

	return &payment,err
}

func (r *paymentRepository) GetByID(ctx context.Context, id uint) (*entity.Payment, error) {
	var payment entity.Payment
	err := r.db.WithContext(ctx).Preload("Booking").First(&payment, id).Error
	return &payment, err
}

func (r *paymentRepository) GetHistoryByUserID(ctx context.Context, userID uint) ([]entity.Payment, error) {
	var payments []entity.Payment
	err := r.db.WithContext(ctx).
		Joins("JOIN bookings ON bookings.id = payments.booking_id").
		Where("bookings.user_id = ?", userID).
		Order("payments.created_at DESC").
		Preload("Booking").
		Find(&payments).Error
	return payments, err
}

func (r *paymentRepository) GetPendingPayment(
	ctx context.Context,
	bookingID uint,
	paymentType string,
) (*entity.Payment, error) {

	var payment entity.Payment

	err := r.db.WithContext(ctx).
		Where(
			"booking_id = ? AND payment_type = ? AND status = ?",
			bookingID,
			paymentType,
			"pending",
		).
		First(&payment).Error

	if err != nil {
		return nil, err
	}

	return &payment, nil
}
