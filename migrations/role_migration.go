package migrations

import (
	"backend/internal/config"
	"backend/internal/entity"
)

func RoleMigration() error {
	return config.DB.AutoMigrate(
		&entity.Role{},
		&entity.UserRole{},
	)
}
