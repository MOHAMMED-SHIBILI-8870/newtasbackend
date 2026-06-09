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

