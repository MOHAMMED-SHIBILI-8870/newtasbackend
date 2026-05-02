package entity

import "time"

type OTP struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Email     string    `gorm:"type:varchar(255);not null" json:"email"`
	OTPCode   string    `gorm:"type:varchar(255);not null" json:"otp_code"`
	Purpose   string    `gorm:"type:varchar(50);not null" json:"purpose"`
	IsUsed    bool      `gorm:"default:false;not null" json:"is_used"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}