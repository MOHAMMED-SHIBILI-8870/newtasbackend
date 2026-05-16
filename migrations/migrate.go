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
		&entity.Trip{},
		&entity.ChatRequest{},
		&entity.ChatResponse{},
	)
}