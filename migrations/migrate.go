package migrations

import (
	"backend/internal/config"
	"backend/internal/entity"
)

func Migrations() error {
	if err := config.DB.AutoMigrate(
		&entity.User{},
		&entity.RefreshToken{},
		&entity.OTP{},
		&entity.ChatResponse{},
		&entity.AITripRequest{},
		&entity.Trip{},
		&entity.TripPlan{},
		&entity.TripSlot{},
		&entity.Booking{},
		&entity.BookingPlan{},
		&entity.Notification{},
		&entity.Payment{},
		&entity.Guide{},
	); err != nil {
		return err
	}

	if err := RoleMigration(); err != nil {
		return err
	}
	if err := PermissionMigration(); err != nil {
		return err
	}
	if err := VehicleMigration(); err != nil {
		return err
	}
	if err := OfferMigration(); err != nil {
		return err
	}
	if err := ReviewMigration(); err != nil {
		return err
	}
	if err := ComplaintMigration(); err != nil {
		return err
	}
	if err := TrackingMigration(); err != nil {
		return err
	}

	return nil
}
