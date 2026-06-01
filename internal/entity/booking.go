package entity

import (
	"time"
)

// Booking links an authenticated User to a master Trip package
type Booking struct {
	ID              uint          `gorm:"primaryKey" json:"id"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	UserID          uint          `gorm:"not null;index" json:"user_id"`
	TripID          uint          `gorm:"not null;index" json:"trip_id"`
	VehicleID       *uint         `gorm:"index" json:"vehicle_id,omitempty"`
	OfferID         *uint         `gorm:"index" json:"offer_id,omitempty"`
	Status          string        `gorm:"type:varchar(50);default:'pending'" json:"status"` // pending, confirmed, cancelled
	SeatsBooked     int           `gorm:"not null;default:1" json:"seats_booked"`
	CouponCode      string        `gorm:"type:varchar(100)" json:"coupon_code,omitempty"`
	DiscountPercent float64       `gorm:"type:decimal(5,2);default:0" json:"discount_percent"`
	BaseAmount      float64       `gorm:"type:decimal(12,2);default:0" json:"base_amount"`
	FinalAmount     float64       `gorm:"type:decimal(12,2);default:0" json:"final_amount"`
	Trip            Trip          `gorm:"foreignKey:TripID" json:"trip"`
	Vehicle         *Vehicle      `gorm:"foreignKey:VehicleID;constraint:OnDelete:SET NULL;" json:"vehicle,omitempty"`
	Offer           *Offer        `gorm:"foreignKey:OfferID;constraint:OnDelete:SET NULL;" json:"offer,omitempty"`
	CustomPlans     []BookingPlan `gorm:"foreignKey:BookingID;constraint:OnDelete:CASCADE;" json:"custom_plans,omitempty"`
}

// BookingPlan mirrors your structural TripPlan fields, isolated for user modifications
type BookingPlan struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	BookingID   uint    `gorm:"not null;index" json:"booking_id"`
	DayNumber   int     `gorm:"not null" json:"day_number"`
	Title       string  `gorm:"type:varchar(255);not null" json:"title"`
	Description string  `gorm:"type:text" json:"description"`
	Location    string  `gorm:"type:varchar(255)" json:"location"`
	StartTime   string  `gorm:"type:varchar(50)" json:"start_time"`
	EndTime     string  `gorm:"type:varchar(50)" json:"end_time"`
	Category    string  `gorm:"type:varchar(50)" json:"category"`
	Cost        float64 `gorm:"type:decimal(10,2)" json:"cost"`
}

// UpdateBookingPlanInput handles array modifications sent from client side updates
type UpdateBookingPlanInput struct {
	Plans []BookingPlan `json:"plans"`
}

type BookingResponse struct {
	ID              uint          `json:"id"`
	Status          string        `json:"status"`
	UserID          uint          `json:"user_id"`
	TripID          uint          `json:"trip_id"`
	VehicleID       *uint         `json:"vehicle_id,omitempty"`
	OfferID         *uint         `json:"offer_id,omitempty"`
	SeatsBooked     int           `json:"seats_booked"`
	CouponCode      string        `json:"coupon_code,omitempty"`
	DiscountPercent float64       `json:"discount_percent"`
	BaseAmount      float64       `json:"base_amount"`
	FinalAmount     float64       `json:"final_amount"`
	CreatedAt       time.Time     `json:"created_at"`
	CustomPlans     []BookingPlan `json:"custom_plans"`
}
