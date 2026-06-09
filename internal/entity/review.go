package entity

import "time"

type Review struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uint      `gorm:"not null;index;uniqueIndex:idx_user_trip_review" json:"user_id"`
	TripID    uint      `gorm:"not null;index;uniqueIndex:idx_user_trip_review" json:"trip_id"`
	Rating    int       `gorm:"not null;index" json:"rating"`
	Comment   string    `gorm:"type:text" json:"comment,omitempty"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"-"`
	Trip      Trip      `gorm:"foreignKey:TripID;constraint:OnDelete:CASCADE;" json:"-"`
}
