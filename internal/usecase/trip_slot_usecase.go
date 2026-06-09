package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TripSlotUsecase struct {
	slotRepo    repository.TripSlotRepository
	tripRepo    repository.TripRepository
	vehicleRepo repository.VehicleRepository
	db          *gorm.DB
}

func NewTripSlotUsecase(
	slotRepo repository.TripSlotRepository,
	tripRepo repository.TripRepository,
	vehicleRepo repository.VehicleRepository,
	db *gorm.DB,
) *TripSlotUsecase {
	return &TripSlotUsecase{
		slotRepo:    slotRepo,
		tripRepo:    tripRepo,
		vehicleRepo: vehicleRepo,
		db:          db,
	}
}

func normalizeTripSlotStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "scheduled":
		return "scheduled"
	case "active":
		return "active"
	case "fully_booked":
		return "fully_booked"
	case "cancelled":
		return "cancelled"
	case "completed":
		return "completed"
	default:
		return "scheduled"
	}
}

func normalizeBookingType(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "shared":
		return "shared", nil
	case "private":
		return "private", nil
	default:
		return "", errors.New("invalid booking type")
	}
}

func normalizeSlotSeats(slot *entity.TripSlot) error {
	if slot.TotalSeats <= 0 {
		return errors.New("total seats must be greater than zero")
	}
	if slot.BookedSeats < 0 {
		return errors.New("booked seats cannot be negative")
	}
	if slot.AvailableSeats < 0 {
		return errors.New("available seats cannot be negative")
	}
	if slot.BookedSeats > slot.TotalSeats {
		return errors.New("booked seats cannot exceed total seats")
	}

	if strings.EqualFold(slot.Status, "fully_booked") {
		slot.BookedSeats = slot.TotalSeats
		slot.AvailableSeats = 0
		return nil
	}

	slot.AvailableSeats = slot.TotalSeats - slot.BookedSeats
	if slot.AvailableSeats < 0 {
		return errors.New("available seats cannot be negative")
	}

	if slot.AvailableSeats == 0 && slot.BookedSeats == 0 {
		slot.AvailableSeats = slot.TotalSeats
	}

	return nil
}

func ensureSlotAssignmentOverlap(
	tx *gorm.DB,
	column string,
	assignmentID uint,
	startDate time.Time,
	endDate time.Time,
	excludeID uint,
) error {
	if assignmentID == 0 {
		return nil
	}

	var count int64
	if err := tx.Model(&entity.TripSlot{}).
		Where(fmt.Sprintf("%s = ?", column), assignmentID).
		Where("id <> ?", excludeID).
		Where("status NOT IN ?", []string{"cancelled", "completed"}).
		Where("start_date < ? AND end_date > ?", endDate, startDate).
		Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		switch column {
		case "vehicle_id":
			return errors.New("vehicle is already assigned to an overlapping slot")
		case "guide_id":
			return errors.New("guide is already assigned to an overlapping slot")
		case "driver_id":
			return errors.New("driver is already assigned to an overlapping slot")
		default:
			return errors.New("assignment overlaps with an existing slot")
		}
	}

	return nil
}

func (u *TripSlotUsecase) validateVehicle(ctx context.Context, tx *gorm.DB, tripID uint, vehicleID *uint, startDate, endDate time.Time, excludeID uint) error {
	if vehicleID == nil || *vehicleID == 0 {
		return nil
	}

	var vehicle entity.Vehicle
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&vehicle, *vehicleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("vehicle not found")
		}
		return err
	}

	if vehicle.TripID != nil && *vehicle.TripID != tripID {
		return errors.New("vehicle is assigned to another trip")
	}

	return ensureSlotAssignmentOverlap(tx, "vehicle_id", *vehicleID, startDate, endDate, excludeID)
}

func (u *TripSlotUsecase) validateTripAndAssignments(
	tx *gorm.DB,
	slot *entity.TripSlot,
	excludeID uint,
) error {
	var trip entity.Trip
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Plans").
		First(&trip, slot.TripID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("trip not found")
		}
		return err
	}

	if err := u.validateVehicle(tx.Statement.Context, tx, slot.TripID, slot.VehicleID, slot.StartDate, slot.EndDate, excludeID); err != nil {
		return err
	}

	if slot.GuideID != nil && *slot.GuideID != 0 {
		if err := ensureSlotAssignmentOverlap(tx, "guide_id", *slot.GuideID, slot.StartDate, slot.EndDate, excludeID); err != nil {
			return err
		}
	}

	if slot.DriverID != nil && *slot.DriverID != 0 {
		if err := ensureSlotAssignmentOverlap(tx, "driver_id", *slot.DriverID, slot.StartDate, slot.EndDate, excludeID); err != nil {
			return err
		}
	}

	_ = trip
	return nil
}

func (u *TripSlotUsecase) CreateSlot(ctx context.Context, slot *entity.TripSlot) (*entity.TripSlot, error) {
	if u == nil || u.db == nil || u.slotRepo == nil || u.tripRepo == nil || u.vehicleRepo == nil {
		return nil, errors.New("slot service unavailable")
	}

	if slot == nil {
		return nil, errors.New("slot is required")
	}

	if slot.TripID == 0 {
		return nil, errors.New("trip id is required")
	}
	if slot.StartDate.IsZero() {
		return nil, errors.New("start date is required")
	}
	if slot.EndDate.IsZero() {
		return nil, errors.New("end date is required")
	}
	if slot.EndDate.Before(slot.StartDate) {
		return nil, errors.New("end date must be after start date")
	}
	if slot.PriceOverride < 0 {
		return nil, errors.New("price override cannot be negative")
	}

	slot.Status = normalizeTripSlotStatus(slot.Status)

	if err := normalizeSlotSeats(slot); err != nil {
		return nil, err
	}

	if slot.VehicleID != nil && *slot.VehicleID == 0 {
		slot.VehicleID = nil
	}
	if slot.GuideID != nil && *slot.GuideID == 0 {
		slot.GuideID = nil
	}
	if slot.DriverID != nil && *slot.DriverID == 0 {
		slot.DriverID = nil
	}

	var created *entity.TripSlot
	err := u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := u.validateTripAndAssignments(tx, slot, 0); err != nil {
			return err
		}

		if err := u.slotRepo.CreateTx(tx, slot); err != nil {
			return err
		}

		created = slot
		return nil
	})
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (u *TripSlotUsecase) GetSlotByID(ctx context.Context, id uint) (*entity.TripSlot, error) {
	if u == nil || u.slotRepo == nil {
		return nil, errors.New("slot service unavailable")
	}
	if id == 0 {
		return nil, errors.New("slot id is required")
	}
	return u.slotRepo.GetByID(ctx, id)
}

func (u *TripSlotUsecase) ListSlots(ctx context.Context) ([]entity.TripSlot, error) {
	if u == nil || u.slotRepo == nil {
		return nil, errors.New("slot service unavailable")
	}
	return u.slotRepo.List(ctx)
}

func (u *TripSlotUsecase) GetUpcomingSlotsByTripID(ctx context.Context, tripID uint) ([]entity.TripSlot, error) {
	if u == nil || u.slotRepo == nil {
		return nil, errors.New("slot service unavailable")
	}
	if tripID == 0 {
		return nil, errors.New("trip id is required")
	}
	return u.slotRepo.GetUpcomingByTripID(ctx, tripID)
}

func (u *TripSlotUsecase) UpdateSlot(ctx context.Context, id uint, input entity.UpdateTripSlotInput) (*entity.TripSlot, error) {
	if u == nil || u.db == nil || u.slotRepo == nil || u.tripRepo == nil || u.vehicleRepo == nil {
		return nil, errors.New("slot service unavailable")
	}
	if id == 0 {
		return nil, errors.New("slot id is required")
	}

	var updated *entity.TripSlot
	err := u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var slot entity.TripSlot
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Trip").
			Preload("Trip.Plans").
			Preload("Vehicle").
			First(&slot, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("slot not found")
			}
			return err
		}

		if input.TripID != nil {
			if *input.TripID == 0 {
				return errors.New("trip id is required")
			}
			slot.TripID = *input.TripID
		}
		if input.VehicleID != nil {
			if *input.VehicleID == 0 {
				slot.VehicleID = nil
			} else {
				slot.VehicleID = input.VehicleID
			}
		}
		if input.GuideID != nil {
			if *input.GuideID == 0 {
				slot.GuideID = nil
			} else {
				slot.GuideID = input.GuideID
			}
		}
		if input.DriverID != nil {
			if *input.DriverID == 0 {
				slot.DriverID = nil
			} else {
				slot.DriverID = input.DriverID
			}
		}
		if input.StartDate != nil {
			slot.StartDate = *input.StartDate
		}
		if input.EndDate != nil {
			slot.EndDate = *input.EndDate
		}
		if input.TotalSeats != nil {
			slot.TotalSeats = *input.TotalSeats
		}
		if input.AvailableSeats != nil {
			slot.AvailableSeats = *input.AvailableSeats
		}
		if input.BookedSeats != nil {
			slot.BookedSeats = *input.BookedSeats
		}
		if input.PriceOverride != nil {
			slot.PriceOverride = *input.PriceOverride
		}
		if input.Status != nil {
			slot.Status = *input.Status
		}

		if slot.EndDate.Before(slot.StartDate) {
			return errors.New("end date must be after start date")
		}
		if slot.PriceOverride < 0 {
			return errors.New("price override cannot be negative")
		}

		slot.Status = normalizeTripSlotStatus(slot.Status)

		if err := normalizeSlotSeats(&slot); err != nil {
			return err
		}

		if err := u.validateTripAndAssignments(tx, &slot, slot.ID); err != nil {
			return err
		}

		if err := u.slotRepo.UpdateTx(tx, &slot); err != nil {
			return err
		}

		updated = &slot
		return nil
	})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (u *TripSlotUsecase) DeleteSlot(ctx context.Context, id uint) error {
	if u == nil || u.slotRepo == nil {
		return errors.New("slot service unavailable")
	}
	if id == 0 {
		return errors.New("slot id is required")
	}
	return u.slotRepo.Delete(ctx, id)
}
