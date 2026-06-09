package entity

import "time"

type Vehicle struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	AgencyID       uint      `gorm:"not null;index" json:"agency_id"`
	Name           string    `gorm:"size:150;not null" json:"name"`
	Type           string    `gorm:"size:50;not null;index" json:"type"`
	TotalSeats     int       `gorm:"not null" json:"total_seats"`
	AvailableSeats int       `gorm:"not null" json:"available_seats"`
	PricePerPerson float64   `gorm:"type:decimal(12,2);not null;default:0" json:"price_per_person"`
	Status         string    `gorm:"size:30;not null;default:'active';index" json:"status"`
	TripID         *uint     `gorm:"index;uniqueIndex:idx_vehicle_trip" json:"trip_id,omitempty"`
	Agency         User      `gorm:"foreignKey:AgencyID;constraint:OnDelete:CASCADE;" json:"-"`
	Trip           *Trip     `gorm:"foreignKey:TripID;constraint:OnDelete:SET NULL;" json:"-"`
}
