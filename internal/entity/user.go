package entity

import (
	"time"
)

type User struct {
	ID            uint           `gorm:"primaryKey:autoIncrement" json:"id"`
	FullName      string         `gorm:"size:50;not null" json:"full_name"`
	Email         string         `gorm:"size:50;uniqueIndex;not null" json:"email"`
	HashPassword  string         `gorm:"size:255" json:"-"`
	Role          string         `gorm:"size:30;default:user;not null" json:"role"`
	IsBlocked     bool           `gorm:"default:false;not null" json:"is_blocked"`
	IsVerified    bool           `gorm:"column:is_verified;default:false;not null" json:"is_verified"`
	RefreshTokens []RefreshToken `gorm:"foreignKey:UserId"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
