package migrations

import (
	"backend/internal/config"
	"backend/internal/entity"
)

func OfferMigration() error {
	return config.DB.AutoMigrate(
		&entity.Offer{},
	)
}
