package repository

import (
	"backend/internal/entity"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type VehicleRepository interface {
	Create(ctx context.Context, vehicle *entity.Vehicle) error
	Update(ctx context.Context, vehicle *entity.Vehicle) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*entity.Vehicle, error)
	GetByTripID(ctx context.Context, tripID uint) (*entity.Vehicle, error)
	GetByAgencyID(ctx context.Context, agencyID uint) ([]entity.Vehicle, error)
	List(ctx context.Context) ([]entity.Vehicle, error)
	AdjustAvailableSeats(ctx context.Context, vehicleID uint, delta int) error
}

type vehicleRepository struct {
	db *gorm.DB
}

func NewVehicleRepository(db *gorm.DB) VehicleRepository {
	return &vehicleRepository{db: db}
}

func (r *vehicleRepository) Create(ctx context.Context, vehicle *entity.Vehicle) error {
	if vehicle.AvailableSeats <= 0 {
		vehicle.AvailableSeats = vehicle.TotalSeats
	}
	return r.db.WithContext(ctx).Create(vehicle).Error
}

func (r *vehicleRepository) Update(ctx context.Context, vehicle *entity.Vehicle) error {
	return r.db.WithContext(ctx).Save(vehicle).Error
}

func (r *vehicleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Vehicle{}, id).Error
}

func (r *vehicleRepository) GetByID(ctx context.Context, id uint) (*entity.Vehicle, error) {
	var vehicle entity.Vehicle
	err := r.db.WithContext(ctx).First(&vehicle, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("vehicle not found")
		}
		return nil, err
	}
	return &vehicle, nil
}

func (r *vehicleRepository) GetByTripID(ctx context.Context, tripID uint) (*entity.Vehicle, error) {
	var vehicle entity.Vehicle
	err := r.db.WithContext(ctx).
		Where("trip_id = ?", tripID).
		First(&vehicle).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &vehicle, nil
}

func (r *vehicleRepository) GetByAgencyID(ctx context.Context, agencyID uint) ([]entity.Vehicle, error) {
	var vehicles []entity.Vehicle
	err := r.db.WithContext(ctx).
		Where("agency_id = ?", agencyID).
		Order("created_at DESC").
		Find(&vehicles).Error
	return vehicles, err
}

func (r *vehicleRepository) List(ctx context.Context) ([]entity.Vehicle, error) {
	var vehicles []entity.Vehicle
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&vehicles).Error
	return vehicles, err
}

func (r *vehicleRepository) AdjustAvailableSeats(ctx context.Context, vehicleID uint, delta int) error {
	query := r.db.WithContext(ctx).Model(&entity.Vehicle{}).Where("id = ?", vehicleID)
	if delta < 0 {
		query = query.Where("available_seats >= ?", -delta)
	}

	result := query.Update("available_seats", gorm.Expr("available_seats + ?", delta))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		if delta < 0 {
			return fmt.Errorf("insufficient available seats")
		}
		return fmt.Errorf("vehicle not found")
	}
	return nil
}
