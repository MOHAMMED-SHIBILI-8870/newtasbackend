package entity

import "time"

type Driver struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Name          string     `gorm:"size:100;not null" json:"name"`
	Email         string     `gorm:"size:100;uniqueIndex;not null" json:"email"`
	Phone         string     `gorm:"size:20;not null" json:"phone"`
	Address       string     `gorm:"type:text" json:"address"`
	LicenseNumber string     `gorm:"size:50;not null" json:"license_number"`
	LicenseExpiry time.Time  `gorm:"type:date;not null" json:"license_expiry"`
	Status        string     `gorm:"size:20;not null;default:'active';index" json:"status"`
	VehicleID     *uint      `gorm:"index" json:"vehicle_id,omitempty"`
	UserID        *uint      `gorm:"index;uniqueIndex" json:"user_id,omitempty"`

	User          *User      `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL;" json:"user,omitempty"`
	Vehicle       *Vehicle   `gorm:"foreignKey:VehicleID;constraint:OnDelete:SET NULL;" json:"vehicle,omitempty"`
}
