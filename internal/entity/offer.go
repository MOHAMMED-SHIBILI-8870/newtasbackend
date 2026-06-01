package entity

import "time"

type Offer struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Code            string    `gorm:"size:100;uniqueIndex;not null" json:"code"`
	Title           string    `gorm:"size:150;not null" json:"title"`
	Description     string    `gorm:"type:text" json:"description,omitempty"`
	DiscountPercent float64   `gorm:"type:decimal(5,2);not null;default:0" json:"discount_percent"`
	ExpiryDate      time.Time `gorm:"not null;index" json:"expiry_date"`
	Active          bool      `gorm:"not null;default:true;index" json:"active"`
}
