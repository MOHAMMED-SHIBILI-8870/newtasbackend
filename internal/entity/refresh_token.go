package entity

import "time"

type RefreshToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
	Token     string    `gorm:"not null;unique"`
	ExpiredAt time.Time `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
