package entity

import "time"

type Booking struct {
    ID            uint      `gorm:"primaryKey" json:"id"`
    UserId        uint      `json:"user_id"`
    TripId        uint      `json:"trip_id"`
    BookingDate   time.Time `json:"booking_date"`
    Status        string    `gorm:"default:'pending'" json:"status"` // pending, confirmed, cancelled
    NumberOfSlots uint      `json:"number_of_slots"`
    TotalPrice    float64   `json:"total_price"`
    
    // Relationships
    User User `gorm:"foreignKey:UserId" json:"-"`
    Trip Trip `gorm:"foreignKey:TripId" json:"trip"`
}