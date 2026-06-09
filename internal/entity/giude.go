package entity

import "time"

type Guide struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      `gorm:"uniqueIndex;not null"`
	Bio         string    `gorm:"type:text"`
	Experience  int
	Languages   string
	IsAvailable bool      `gorm:"default:true"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	User User `gorm:"foreignKey:UserID"`
}