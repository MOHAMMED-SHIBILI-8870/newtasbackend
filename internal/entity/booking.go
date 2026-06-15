package entity

import (
	"time"
)

// Booking links an authenticated User to a master Trip package
type Booking struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	UserID    uint  `gorm:"not null;index;uniqueIndex:idx_booking_user_slot" json:"user_id"`
	User      User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	TripID    uint  `gorm:"not null;index" json:"trip_id"`
	SlotID    *uint `gorm:"index;uniqueIndex:idx_booking_user_slot" json:"slot_id,omitempty"`
	VehicleID *uint `gorm:"index" json:"vehicle_id,omitempty"`
	OfferID   *uint `gorm:"index" json:"offer_id,omitempty"`

	BookingType string `gorm:"type:varchar(20);default:'shared'" json:"booking_type"`
	Status      string `gorm:"type:varchar(50);default:'pending'" json:"status"`

	StartDate *time.Time `gorm:"type:date" json:"start_date,omitempty"`
	EndDate   *time.Time `gorm:"type:date" json:"end_date,omitempty"`

	SeatsBooked     int     `gorm:"not null;default:1" json:"seats_booked"`
	CouponCode      string  `gorm:"type:varchar(100)" json:"coupon_code,omitempty"`
	DiscountPercent float64 `gorm:"type:decimal(5,2);default:0" json:"discount_percent"`

	BaseAmount  float64 `gorm:"type:decimal(12,2);default:0" json:"base_amount"`
	FinalAmount float64 `gorm:"type:decimal(12,2);default:0" json:"final_amount"`

	// Payment Summary
	AdvancePercent float64 `gorm:"type:decimal(5,2);default:20" json:"advance_percent"`
	AdvanceAmount  float64 `gorm:"type:decimal(12,2);default:0" json:"advance_amount"`
	BalanceAmount  float64 `gorm:"type:decimal(12,2);default:0" json:"balance_amount"`

	PaymentStatus  string     `gorm:"type:varchar(50);default:'pending'" json:"payment_status"`
	BalanceDueDate *time.Time `json:"balance_due_date,omitempty"`

	Trip    Trip      `gorm:"foreignKey:TripID" json:"trip"`
	Slot    *TripSlot `gorm:"foreignKey:SlotID;constraint:OnDelete:SET NULL;" json:"slot,omitempty"`
	Vehicle *Vehicle  `gorm:"foreignKey:VehicleID;constraint:OnDelete:SET NULL;" json:"vehicle,omitempty"`
	Offer   *Offer    `gorm:"foreignKey:OfferID;constraint:OnDelete:SET NULL;" json:"offer,omitempty"`

	CustomPlans []BookingPlan `gorm:"foreignKey:BookingID;constraint:OnDelete:CASCADE;" json:"custom_plans,omitempty"`
	Payments    []Payment     `gorm:"foreignKey:BookingID" json:"payments,omitempty"`
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

type UserResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type BookingResponse struct {
	ID              uint          `json:"id"`
	Status          string        `json:"status"`
	UserID          uint          `json:"user_id"`
	User            *UserResponse `json:"user,omitempty"`
	TripID          uint          `json:"trip_id"`
	Trip            *Trip         `json:"trip,omitempty"`
	SlotID          *uint         `json:"slot_id,omitempty"`
	VehicleID       *uint         `json:"vehicle_id,omitempty"`
	OfferID         *uint         `json:"offer_id,omitempty"`
	BookingType     string        `json:"booking_type"`
	SeatsBooked     int           `json:"seats_booked"`
	CouponCode      string        `json:"coupon_code,omitempty"`
	DiscountPercent float64       `json:"discount_percent"`
	BaseAmount      float64       `json:"base_amount"`
	FinalAmount     float64       `json:"final_amount"`
	BalanceAmount   float64       `json:"balance_amount"`
	PaymentStatus   string        `json:"payment_status"`
	CreatedAt       time.Time     `json:"created_at"`
	StartDate       *time.Time    `json:"start_date,omitempty"`
	EndDate         *time.Time    `json:"end_date,omitempty"`
	CustomPlans     []BookingPlan `json:"custom_plans"`
}
