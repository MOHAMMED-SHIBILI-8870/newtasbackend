package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type VehicleUsecase struct {
	vehicleRepo repository.VehicleRepository
	tripRepo    repository.TripRepository
	userRepo    repository.UserRepository
	db          *gorm.DB
}

func NewVehicleUsecase(
	vehicleRepo repository.VehicleRepository,
	tripRepo repository.TripRepository,
	userRepo repository.UserRepository,
	db *gorm.DB,
) *VehicleUsecase {
	return &VehicleUsecase{
		vehicleRepo: vehicleRepo,
		tripRepo:    tripRepo,
		userRepo:    userRepo,
		db:          db,
	}
}

func normalizeVehicleType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "bus":
		return "Bus"
	case "car":
		return "Car"
	case "traveler":
		return "Traveler"
	default:
		return ""
	}
}

func normalizeVehicleStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "inactive":
		return "inactive"
	case "maintenance":
		return "maintenance"
	case "assigned":
		return "assigned"
	default:
		return "active"
	}
}

func (u *VehicleUsecase) CreateVehicle(ctx context.Context, actorID uint, actorRole string, vehicle *entity.Vehicle) error {
	if vehicle == nil {
		return errors.New("vehicle is required")
	}

	vehicle.Name = strings.TrimSpace(vehicle.Name)
	vehicle.Type = normalizeVehicleType(vehicle.Type)
	vehicle.Status = normalizeVehicleStatus(vehicle.Status)

	if vehicle.Name == "" {
		return errors.New("vehicle name is required")
	}
	if vehicle.Type == "" {
		return errors.New("invalid vehicle type")
	}
	if vehicle.TotalSeats <= 0 {
		return errors.New("total seats must be greater than zero")
	}
	if vehicle.AvailableSeats <= 0 || vehicle.AvailableSeats > vehicle.TotalSeats {
		vehicle.AvailableSeats = vehicle.TotalSeats
	}

	if NormalizeRole(actorRole) != "admin" {
		vehicle.AgencyID = actorID
	}
	if vehicle.AgencyID == 0 {
		return errors.New("agency id is required")
	}

	return u.vehicleRepo.Create(ctx, vehicle)
}

func (u *VehicleUsecase) UpdateVehicle(ctx context.Context, actorID uint, actorRole string, id uint, vehicle *entity.Vehicle) error {
	if id == 0 {
		return errors.New("vehicle id is required")
	}
	if vehicle == nil {
		return errors.New("vehicle is required")
	}

	existing, err := u.vehicleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("vehicle not found")
	}

	if NormalizeRole(actorRole) != "admin" && existing.AgencyID != actorID {
		return errors.New("access denied")
	}

	if strings.TrimSpace(vehicle.Name) != "" {
		existing.Name = strings.TrimSpace(vehicle.Name)
	}
	if normalized := normalizeVehicleType(vehicle.Type); normalized != "" {
		existing.Type = normalized
	}
	if vehicle.TotalSeats > 0 {
		existing.TotalSeats = vehicle.TotalSeats
	}
	if vehicle.AvailableSeats >= 0 {
		existing.AvailableSeats = vehicle.AvailableSeats
	}
	if vehicle.PricePerPerson >= 0 {
		existing.PricePerPerson = vehicle.PricePerPerson
	}
	if strings.TrimSpace(vehicle.Status) != "" {
		existing.Status = normalizeVehicleStatus(vehicle.Status)
	}
	if NormalizeRole(actorRole) == "admin" && vehicle.AgencyID > 0 {
		existing.AgencyID = vehicle.AgencyID
	}
	if vehicle.TripID != nil {
		existing.TripID = vehicle.TripID
	}

	if existing.AvailableSeats > existing.TotalSeats {
		existing.AvailableSeats = existing.TotalSeats
	}
	if existing.AvailableSeats < 0 {
		existing.AvailableSeats = 0
	}

	return u.vehicleRepo.Update(ctx, existing)
}

func (u *VehicleUsecase) DeleteVehicle(ctx context.Context, actorID uint, actorRole string, id uint) error {
	if id == 0 {
		return errors.New("vehicle id is required")
	}

	existing, err := u.vehicleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("vehicle not found")
	}

	if NormalizeRole(actorRole) != "admin" && existing.AgencyID != actorID {
		return errors.New("access denied")
	}

	return u.vehicleRepo.Delete(ctx, id)
}

func (u *VehicleUsecase) ListVehicles(ctx context.Context, actorID uint, actorRole string) ([]entity.Vehicle, error) {
	role := NormalizeRole(actorRole)
	switch role {
	case "admin":
		return u.vehicleRepo.List(ctx)
	case "agency":
		return u.vehicleRepo.GetByAgencyID(ctx, actorID)
	default:
		vehicles, err := u.vehicleRepo.List(ctx)
		if err != nil {
			return nil, err
		}
		filtered := make([]entity.Vehicle, 0, len(vehicles))
		for _, vehicle := range vehicles {
			if vehicle.Status == "active" || vehicle.Status == "assigned" {
				filtered = append(filtered, vehicle)
			}
		}
		return filtered, nil
	}
}

func (u *VehicleUsecase) GetVehicleByID(ctx context.Context, id uint) (*entity.Vehicle, error) {
	if id == 0 {
		return nil, errors.New("vehicle id is required")
	}
	return u.vehicleRepo.GetByID(ctx, id)
}

func (u *VehicleUsecase) AssignVehicleToTrip(ctx context.Context, actorID uint, actorRole string, vehicleID uint, tripID uint) error {
	if vehicleID == 0 || tripID == 0 {
		return errors.New("vehicle and trip ids are required")
	}
	if u.db == nil {
		return errors.New("database unavailable")
	}

	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var vehicle entity.Vehicle
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&vehicle, vehicleID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("vehicle not found")
			}
			return err
		}

		if NormalizeRole(actorRole) != "admin" && vehicle.AgencyID != actorID {
			return errors.New("access denied")
		}

		var trip entity.Trip
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&trip, tripID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("trip not found")
			}
			return err
		}

		var existingTripVehicle entity.Vehicle
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("trip_id = ? AND id <> ?", tripID, vehicleID).
			First(&existingTripVehicle).Error
		if err == nil {
			return errors.New("trip already has an assigned vehicle")
		}

		if vehicle.TripID != nil && *vehicle.TripID != tripID {
			return errors.New("vehicle is already assigned to another trip")
		}

		vehicle.TripID = &tripID
		vehicle.Status = normalizeVehicleStatus("assigned")

		return tx.Save(&vehicle).Error
	})
}

func (u *VehicleUsecase) GetVehicleByTripID(ctx context.Context, tripID uint) (*entity.Vehicle, error) {
	if tripID == 0 {
		return nil, errors.New("trip id is required")
	}
	return u.vehicleRepo.GetByTripID(ctx, tripID)
}
