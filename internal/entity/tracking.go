package entity

import "time"

type Tracking struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	BookingID uint      `gorm:"not null;index" json:"booking_id"`
	VehicleID *uint     `gorm:"index" json:"vehicle_id,omitempty"`
	DriverID  *uint     `gorm:"index" json:"driver_id,omitempty"`
	Type      string    `gorm:"type:varchar(20);default:'vehicle'" json:"type"`
	Latitude  float64   `gorm:"not null" json:"latitude"`
	Longitude float64   `gorm:"not null" json:"longitude"`
	Booking   Booking   `gorm:"foreignKey:BookingID;constraint:OnDelete:CASCADE;" json:"-"`
	Vehicle   *Vehicle  `gorm:"foreignKey:VehicleID;constraint:OnDelete:CASCADE;" json:"-"`
	Driver    *Driver   `gorm:"foreignKey:DriverID;constraint:OnDelete:CASCADE;" json:"-"`
}
