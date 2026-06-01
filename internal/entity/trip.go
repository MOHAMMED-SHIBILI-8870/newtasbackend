package entity

import (
	"time"

	"gorm.io/gorm"
)

type Trip struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// --- Core Trip Info & Logistics ---
	From        string    `gorm:"type:varchar(255);not null" json:"from"`
	To          string    `gorm:"type:varchar(255);not null" json:"to"`
	StartDate   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"start_date"`
	EndDate     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"end_date"`
	Duration    int       `gorm:"not null;default:1" json:"duration"`
	TripType    string    `gorm:"type:varchar(100);default:'Family'" json:"trip_type"`
	BudgetLevel string    `gorm:"type:varchar(50);default:'Medium'" json:"budget_level"`
	Price       float64   `gorm:"type:decimal(10,2);default:0.00" json:"price"`
	Members     int       `gorm:"default:1" json:"members"`
	Children    int       `gorm:"default:0" json:"children"`
	HotelType   string    `gorm:"type:varchar(100);default:'3 Star'" json:"hotel_type"`
	Transport   string    `gorm:"type:varchar(255);default:'Car'" json:"transport"` // Mixed transport allowed

	// --- Media & Status ---
	ItineraryRaw string `gorm:"type:text" json:"itinerary_raw,omitempty"`
	ImageURL     string `gorm:"type:text" json:"image_url,omitempty"`
	Status       string `gorm:"type:varchar(50);default:'active'" json:"status"`

	// --- Relationships ---
	Plans []TripPlan `gorm:"foreignKey:TripID;constraint:OnDelete:CASCADE;" json:"plans,omitempty"`
}
