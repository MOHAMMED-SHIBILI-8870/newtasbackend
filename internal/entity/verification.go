package entity

import (
	"time"
)

type Verification struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	UserID      uint   `gorm:"not null;index" json:"user_id"`
	BookingID   uint   `gorm:"not null;index" json:"booking_id"`

	FullName    string `gorm:"type:varchar(255);not null" json:"full_name"`
	Address     string `gorm:"type:text;not null" json:"address"`
	PhoneNumber string `gorm:"type:varchar(50);not null" json:"phone_number"`
	IDImageURL  string `gorm:"type:varchar(255)" json:"id_image_url"`
	Members     int    `gorm:"default:1" json:"members"`
	
	Status      string `gorm:"type:varchar(50);default:'pending'" json:"status"`
}
