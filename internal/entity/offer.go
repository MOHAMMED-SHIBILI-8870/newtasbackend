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
	DiscountType    string    `gorm:"type:varchar(20);default:'percentage'" json:"discount_type"`
	FixedDiscount   float64   `gorm:"type:decimal(10,2);default:0" json:"fixed_discount"`
	MaxUsage        int       `gorm:"default:0" json:"max_usage"`
	CurrentUsage    int       `gorm:"default:0" json:"current_usage"`
	TripID          *uint     `gorm:"index" json:"trip_id,omitempty"`
	ExpiryDate      time.Time `gorm:"not null;index" json:"expiry_date"`
	Active          bool      `gorm:"not null;default:true;index" json:"active"`
	Trip            *Trip     `gorm:"foreignKey:TripID;constraint:OnDelete:SET NULL;" json:"-"`
}
