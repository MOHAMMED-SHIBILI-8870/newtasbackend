package migrations

import (
	"backend/internal/config"
	"backend/internal/entity"
)

func TrackingMigration() error {
	return config.DB.AutoMigrate(
		&entity.Tracking{},
	)
}
