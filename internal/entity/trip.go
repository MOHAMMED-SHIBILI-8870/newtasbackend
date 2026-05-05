package entity

import (
	"time"

	"gorm.io/gorm"
)

// Trip represents the core data model for a travel plan
type Trip struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // Soft delete for production safety

	// Business Fields
	Destination string  `gorm:"type:varchar(255);not null" json:"destination"`
	Budget      float64 `gorm:"type:decimal(10,2)" json:"budget"`
	Duration    int     `json:"duration"` // Number of days
	Description string  `gorm:"type:text" json:"description"`
	ImageUrl    string  `json:"image_url"`

	// Relationship  Every trip belongs to a User
	UserId uint `gorm:"not null" json:"user_id"`
	User   User `gorm:"foreignKey:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
}
