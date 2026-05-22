package entity

import "time"

type Notification struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	UserID           uint           `gorm:"not null;index" json:"user_id"`
	Type             string         `gorm:"type:varchar(50);not null;default:'general';index" json:"type"`
	Title            string         `gorm:"type:varchar(150);not null" json:"title"`
	Message          string         `gorm:"type:text;not null" json:"message"`
	BookingID        *uint          `gorm:"index" json:"booking_id,omitempty"`
	AITripRequestID  *uint          `gorm:"index" json:"ai_trip_request_id,omitempty"`
	IsRead           bool           `gorm:"not null;default:false;index" json:"is_read"`
	Metadata         string         `gorm:"type:text" json:"metadata,omitempty"`
	User             User           `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Booking          *Booking       `gorm:"foreignKey:BookingID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	AITripRequest    *AITripRequest `gorm:"foreignKey:AITripRequestID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}
