package migrations

import (
	"backend/internal/config"
	"backend/internal/entity"
)

func PermissionMigration() error {
	return config.DB.AutoMigrate(
		&entity.Permission{},
		&entity.RolePermission{},
	)
}
