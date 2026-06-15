package entity

import "time"

// TripSlot represents one scheduled departure of a trip template.
type TripSlot struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	TripID         uint      `gorm:"not null;index" json:"trip_id"`
	VehicleID      *uint     `gorm:"index" json:"vehicle_id,omitempty"`
	GuideID        *uint     `gorm:"index" json:"guide_id,omitempty"`
	DriverID       *uint     `gorm:"index" json:"driver_id,omitempty"`
	StartDate      time.Time `gorm:"not null;index" json:"start_date"`
	EndDate        time.Time `gorm:"not null;index" json:"end_date"`
	TotalSeats     int       `gorm:"not null;default:0" json:"total_seats"`
	AvailableSeats int       `gorm:"not null;default:0" json:"available_seats"`
	BookedSeats    int       `gorm:"not null;default:0" json:"booked_seats"`
	PriceOverride  float64   `gorm:"type:decimal(12,2);default:0" json:"price_override"`
	Status         string    `gorm:"size:30;not null;default:'scheduled';index" json:"status"`
	Trip           Trip      `gorm:"foreignKey:TripID;constraint:OnDelete:CASCADE;" json:"trip"`
	Vehicle        *Vehicle  `gorm:"foreignKey:VehicleID;constraint:OnDelete:SET NULL;" json:"vehicle,omitempty"`
	Driver         *Driver   `gorm:"foreignKey:DriverID;constraint:OnDelete:SET NULL;" json:"driver,omitempty"`
}

// UpdateTripSlotInput supports partial admin updates.
type UpdateTripSlotInput struct {
	TripID         *uint      `json:"trip_id"`
	VehicleID      *uint      `json:"vehicle_id"`
	GuideID        *uint      `json:"guide_id"`
	DriverID       *uint      `json:"driver_id"`
	StartDate      *time.Time `json:"start_date"`
	EndDate        *time.Time `json:"end_date"`
	TotalSeats     *int       `json:"total_seats"`
	AvailableSeats *int       `json:"available_seats"`
	BookedSeats    *int       `json:"booked_seats"`
	PriceOverride  *float64   `json:"price_override"`
	Status         *string    `json:"status"`
}
