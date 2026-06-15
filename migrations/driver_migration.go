package migrations

import (
	"backend/internal/config"
	"backend/internal/entity"
)

func DriverMigration() error {
	return config.DB.AutoMigrate(
		&entity.Driver{},
	)
}
