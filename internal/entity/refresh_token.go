package entity

import "time"

type RefreshToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserId    uint      `gorm:"not null;index"`
	Token     string    `gorm:"not null;unique"`
	ExpiredAt time.Time `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}