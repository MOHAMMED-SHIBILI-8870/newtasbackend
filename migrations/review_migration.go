package migrations

import (
	"backend/internal/config"
	"backend/internal/entity"
)

func ReviewMigration() error {
	return config.DB.AutoMigrate(
		&entity.Review{},
	)
}
