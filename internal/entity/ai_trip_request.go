package entity

import (
	"time"

	"gorm.io/gorm"
)

const (
	AITripStatusPending  = "pending"
	AITripStatusApproved = "approved"
	AITripStatusRejected = "rejected"
)

type AITripRequest struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	UserID         uint           `gorm:"not null;index" json:"user_id"`
	From           string         `gorm:"type:varchar(255);not null" json:"from"`
	To             string         `gorm:"type:varchar(255);not null" json:"to"`
	Days           int            `gorm:"not null;default:1" json:"days"`
	TripType       string         `gorm:"type:varchar(100);default:'Family'" json:"trip_type"`
	BudgetLevel    string         `gorm:"type:varchar(50);default:'Medium'" json:"budget_level"`
	Members        int            `gorm:"not null;default:1" json:"members"`
	Children       int            `gorm:"not null;default:0" json:"children"`
	HotelType      string         `gorm:"type:varchar(100);default:'3 Star'" json:"hotel_type"`
	Transport      string         `gorm:"type:varchar(255);default:'Car'" json:"transport"`
	Prompt         string         `gorm:"type:text;not null" json:"prompt"`
	GeneratedPlan  string         `gorm:"type:text;not null" json:"generated_plan"`
	Status         string         `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`
	AdminNote      string         `gorm:"type:text" json:"admin_note,omitempty"`
	TripID         *uint          `gorm:"index" json:"trip_id,omitempty"`
	ReviewedByID   *uint          `gorm:"index" json:"reviewed_by_id,omitempty"`
	ReviewedAt     *time.Time      `json:"reviewed_at,omitempty"`
	User           User           `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Trip           *Trip          `gorm:"foreignKey:TripID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	ReviewedBy     *User          `gorm:"foreignKey:ReviewedByID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
}

type AITripRequestInput struct {
	From          string `json:"from"`
	To            string `json:"to"`
	Days          int    `json:"days"`
	TripType      string `json:"trip_type"`
	BudgetLevel   string `json:"budget_level"`
	Members       int    `json:"members"`
	Children      int    `json:"children"`
	HotelType     string `json:"hotel_type"`
	Transport     string `json:"transport"`
	Prompt        string `json:"prompt"`
	GeneratedPlan string `json:"generated_plan"`
}

type AITripReviewInput struct {
	AdminNote string `json:"admin_note"`
}
