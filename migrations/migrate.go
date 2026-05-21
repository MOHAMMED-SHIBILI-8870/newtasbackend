package migrations

import (
	"backend/internal/config"
	"backend/internal/entity"
)

func Migrations() error {
	return config.DB.AutoMigrate(
		&entity.User{},
		&entity.RefreshToken{},
		&entity.OTP{},
		&entity.ChatResponse{},
		&entity.Trip{},
		&entity.TripPlan{},
		&entity.Booking{},
		&entity.BookingPlan{},
	)
}