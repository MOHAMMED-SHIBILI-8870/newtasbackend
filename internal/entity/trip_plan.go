package entity

import (
	"time"

	"gorm.io/gorm"
)

type TripPlan struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	TripID      uint   `gorm:"not null;index" json:"trip_id"`
	DayNumber   int    `gorm:"not null" json:"day_number"`
	Title       string `gorm:"type:varchar(255);not null" json:"title"`
	Description string `gorm:"type:text" json:"description"`
	Location    string `gorm:"type:varchar(255)" json:"location"`
	StartTime   string `gorm:"type:varchar(50)" json:"start_time"`
	EndTime     string `gorm:"type:varchar(50)" json:"end_time"`

	Category string  `gorm:"type:varchar(50)" json:"category"`
	Cost     float64 `gorm:"type:decimal(10,2)" json:"cost"`

	Trip Trip `gorm:"foreignKey:TripID" json:"-"`
}
