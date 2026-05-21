package entity

import (
	"time"
)

// Booking links an authenticated User to a master Trip package
type Booking struct {
	ID          uint          `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	UserID      uint          `gorm:"not null;index" json:"user_id"`
	TripID      uint          `gorm:"not null;index" json:"trip_id"`
	Status      string        `gorm:"type:varchar(50);default:'pending'" json:"status"` // pending, confirmed, cancelled
	Trip        Trip          `gorm:"foreignKey:TripID" json:"trip"`
	CustomPlans []BookingPlan `gorm:"foreignKey:BookingID;constraint:OnDelete:CASCADE;" json:"custom_plans,omitempty"`
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