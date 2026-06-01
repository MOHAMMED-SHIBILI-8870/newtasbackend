package migrations

import (
	"backend/internal/config"
	"backend/internal/entity"
)

func ComplaintMigration() error {
	return config.DB.AutoMigrate(
		&entity.Complaint{},
	)
}
