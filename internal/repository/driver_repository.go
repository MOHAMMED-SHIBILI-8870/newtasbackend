package repository

import (
	"backend/internal/entity"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type DriverRepository interface {
	Create(ctx context.Context, driver *entity.Driver) error
	Update(ctx context.Context, driver *entity.Driver) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*entity.Driver, error)
	GetByUserID(ctx context.Context, userID uint) (*entity.Driver, error)
	List(ctx context.Context) ([]entity.Driver, error)
}

type driverRepository struct {
	db *gorm.DB
}

func NewDriverRepository(db *gorm.DB) DriverRepository {
	return &driverRepository{db: db}
}

func (r *driverRepository) Create(ctx context.Context, driver *entity.Driver) error {
	return r.db.WithContext(ctx).Create(driver).Error
}

func (r *driverRepository) Update(ctx context.Context, driver *entity.Driver) error {
	return r.db.WithContext(ctx).Save(driver).Error
}

func (r *driverRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var driver entity.Driver
		if err := tx.First(&driver, id).Error; err != nil {
			return err
		}

		// Set Vehicle's DriverID to nil
		if err := tx.Model(&entity.Vehicle{}).Where("driver_id = ?", driver.ID).Update("driver_id", nil).Error; err != nil {
			return err
		}

		// Delete User if associated
		if driver.UserID != nil {
			if err := tx.Delete(&entity.User{}, *driver.UserID).Error; err != nil {
				return err
			}
		}

		return tx.Delete(&driver).Error
	})
}

func (r *driverRepository) GetByID(ctx context.Context, id uint) (*entity.Driver, error) {
	var driver entity.Driver
	err := r.db.WithContext(ctx).Preload("Vehicle").First(&driver, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("driver not found")
		}
		return nil, err
	}
	return &driver, nil
}

func (r *driverRepository) GetByUserID(ctx context.Context, userID uint) (*entity.Driver, error) {
	var driver entity.Driver
	err := r.db.WithContext(ctx).Preload("Vehicle").Where("user_id = ?", userID).First(&driver).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("driver profile not found")
		}
		return nil, err
	}
	return &driver, nil
}

func (r *driverRepository) List(ctx context.Context) ([]entity.Driver, error) {
	var drivers []entity.Driver
	err := r.db.WithContext(ctx).Preload("Vehicle").Order("created_at DESC").Find(&drivers).Error
	return drivers, err
}
