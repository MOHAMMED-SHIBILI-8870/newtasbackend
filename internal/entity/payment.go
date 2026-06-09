package entity

import "time"

type Payment struct {
	ID uint `gorm:"primaryKey" json:"id"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	BookingID uint `gorm:"not null;index" json:"booking_id"`

	PaymentType string `gorm:"type:varchar(20);not null" json:"payment_type"`

	Status string `gorm:"type:varchar(20);default:'pending'" json:"status"`

	Amount float64 `gorm:"type:decimal(12,2);not null" json:"amount"`

	RazorpayOrderID   string `gorm:"type:varchar(255);index" json:"razorpay_order_id"`
	RazorpayPaymentID string `gorm:"type:varchar(255);index" json:"razorpay_payment_id"`
	RazorpaySignature string `gorm:"type:varchar(255)" json:"-"`

	Booking Booking `gorm:"foreignKey:BookingID" json:"-"`
}
