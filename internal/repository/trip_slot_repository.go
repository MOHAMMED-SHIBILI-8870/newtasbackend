package repository

import (
	"backend/internal/entity"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type TripSlotRepository interface {
	Create(ctx context.Context, slot *entity.TripSlot) error
	CreateTx(tx *gorm.DB, slot *entity.TripSlot) error
	Update(ctx context.Context, slot *entity.TripSlot) error
	UpdateTx(tx *gorm.DB, slot *entity.TripSlot) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*entity.TripSlot, error)
	GetByTripID(ctx context.Context, tripID uint) ([]entity.TripSlot, error)
	GetUpcomingByTripID(ctx context.Context, tripID uint) ([]entity.TripSlot, error)
	List(ctx context.Context) ([]entity.TripSlot, error)
}

type tripSlotRepository struct {
	db *gorm.DB
}

func NewTripSlotRepository(db *gorm.DB) TripSlotRepository {
	return &tripSlotRepository{db: db}
}

func (r *tripSlotRepository) Create(ctx context.Context, slot *entity.TripSlot) error {
	return r.db.WithContext(ctx).Create(slot).Error
}

func (r *tripSlotRepository) CreateTx(tx *gorm.DB, slot *entity.TripSlot) error {
	return tx.Create(slot).Error
}

func (r *tripSlotRepository) Update(ctx context.Context, slot *entity.TripSlot) error {
	return r.db.WithContext(ctx).Save(slot).Error
}

func (r *tripSlotRepository) UpdateTx(tx *gorm.DB, slot *entity.TripSlot) error {
	return tx.Save(slot).Error
}

func (r *tripSlotRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.TripSlot{}, id).Error
}

func (r *tripSlotRepository) GetByID(ctx context.Context, id uint) (*entity.TripSlot, error) {
	var slot entity.TripSlot
	err := r.db.WithContext(ctx).
		Preload("Trip").
		Preload("Trip.Plans").
		Preload("Vehicle").
		First(&slot, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("slot not found")
		}
		return nil, err
	}
	return &slot, nil
}

func (r *tripSlotRepository) GetByTripID(ctx context.Context, tripID uint) ([]entity.TripSlot, error) {
	var slots []entity.TripSlot
	err := r.db.WithContext(ctx).
		Preload("Trip").
		Preload("Trip.Plans").
		Preload("Vehicle").
		Where("trip_id = ?", tripID).
		Order("start_date ASC, created_at DESC").
		Find(&slots).Error
	return slots, err
}

func (r *tripSlotRepository) GetUpcomingByTripID(ctx context.Context, tripID uint) ([]entity.TripSlot, error) {
	var slots []entity.TripSlot
	err := r.db.WithContext(ctx).
		Preload("Trip").
		Preload("Trip.Plans").
		Preload("Vehicle").
		Where("trip_id = ?", tripID).
		Where("status NOT IN ?", []string{"cancelled", "completed"}).
		Where("end_date >= ?", time.Now()).
		Order("start_date ASC, created_at DESC").
		Find(&slots).Error
	return slots, err
}

func (r *tripSlotRepository) List(ctx context.Context) ([]entity.TripSlot, error) {
	var slots []entity.TripSlot
	err := r.db.WithContext(ctx).
		Preload("Trip").
		Preload("Trip.Plans").
		Preload("Vehicle").
		Order("start_date ASC, created_at DESC").
		Find(&slots).Error
	return slots, err
}

func isSlotAssignmentColumn(column string) bool {
	switch strings.ToLower(strings.TrimSpace(column)) {
	case "vehicle_id", "guide_id", "driver_id":
		return true
	default:
		return false
	}
}
