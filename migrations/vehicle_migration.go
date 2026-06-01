package migrations

import (
	"backend/internal/config"
	"backend/internal/entity"
)

func VehicleMigration() error {
	return config.DB.AutoMigrate(
		&entity.Vehicle{},
	)
}
