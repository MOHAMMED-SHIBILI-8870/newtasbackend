package usecase

import (
	"backend/internal/dto"
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type DriverUsecase struct {
	driverRepo   repository.DriverRepository
	userRepo     repository.UserRepository
	roleRepo     repository.RoleRepository
	vehicleRepo  repository.VehicleRepository
	bookingRepo  repository.BookingRepository
	trackingRepo repository.TrackingRepository
	db           *gorm.DB
}

func NewDriverUsecase(
	driverRepo repository.DriverRepository,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	vehicleRepo repository.VehicleRepository,
	bookingRepo repository.BookingRepository,
	trackingRepo repository.TrackingRepository,
	db *gorm.DB,
) *DriverUsecase {
	return &DriverUsecase{
		driverRepo:   driverRepo,
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		vehicleRepo:  vehicleRepo,
		bookingRepo:  bookingRepo,
		trackingRepo: trackingRepo,
		db:           db,
	}
}

// 📄 Admin: Create Driver
func (u *DriverUsecase) CreateDriver(ctx context.Context, req dto.DriverCreateRequest) (*entity.Driver, error) {
	if req.Email == "" || req.Name == "" || req.Password == "" {
		return nil, errors.New("name, email and password are required")
	}

	// Check email
	existing, _ := u.userRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.New("email already exists")
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	role, err := u.roleRepo.GetByName(ctx, "driver")
	if err != nil || role == nil {
		return nil, errors.New("driver role not configured in DB")
	}

	var driver entity.Driver

	err = u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create User
		user := entity.User{
			FullName:     req.Name,
			Email:        req.Email,
			HashPassword: hash,
			Role:         "driver",
			IsVerified:   true,
			IsBlocked:    false,
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		if err := syncPrimaryRoleTx(tx, user.ID, role.ID, role.Name); err != nil {
			return err
		}

		// Create Driver profile
		driver = entity.Driver{
			Name:          req.Name,
			Email:         req.Email,
			Phone:         req.Phone,
			Address:       req.Address,
			LicenseNumber: req.LicenseNumber,
			LicenseExpiry: req.LicenseExpiry,
			Status:        "active",
			UserID:        &user.ID,
		}
		if err := tx.Create(&driver).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Retrieve fully populated
	return u.driverRepo.GetByID(ctx, driver.ID)
}

// 📄 Admin: List Drivers
func (u *DriverUsecase) ListDrivers(ctx context.Context) ([]entity.Driver, error) {
	return u.driverRepo.List(ctx)
}

// 📄 Admin: Get Driver By ID
func (u *DriverUsecase) GetDriverByID(ctx context.Context, id uint) (*entity.Driver, error) {
	return u.driverRepo.GetByID(ctx, id)
}

// 📄 Admin: Update Driver
func (u *DriverUsecase) UpdateDriver(ctx context.Context, id uint, req dto.DriverUpdateRequest) (*entity.Driver, error) {
	driver, err := u.driverRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	err = u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if req.Name != "" {
			driver.Name = req.Name
			if driver.UserID != nil {
				tx.Model(&entity.User{}).Where("id = ?", *driver.UserID).Update("full_name", req.Name)
			}
		}
		if req.Email != "" && req.Email != driver.Email {
			var existing entity.User
			if err := tx.Where("email = ? AND id <> ?", req.Email, *driver.UserID).First(&existing).Error; err == nil {
				return errors.New("email already in use")
			}
			driver.Email = req.Email
			if driver.UserID != nil {
				tx.Model(&entity.User{}).Where("id = ?", *driver.UserID).Update("email", req.Email)
			}
		}
		if req.Phone != "" {
			driver.Phone = req.Phone
		}
		if req.Address != "" {
			driver.Address = req.Address
		}
		if req.LicenseNumber != "" {
			driver.LicenseNumber = req.LicenseNumber
		}
		if req.LicenseExpiry != nil {
			driver.LicenseExpiry = *req.LicenseExpiry
		}
		if req.Status != "" {
			driver.Status = req.Status
		}
		if req.VehicleID != nil {
			if *req.VehicleID == 0 {
				driver.VehicleID = nil
			} else {
				driver.VehicleID = req.VehicleID
			}
		}

		if err := tx.Save(driver).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return u.driverRepo.GetByID(ctx, driver.ID)
}

// 📄 Admin: Delete Driver
func (u *DriverUsecase) DeleteDriver(ctx context.Context, id uint) error {
	return u.driverRepo.Delete(ctx, id)
}

// 📄 Admin: Assign Vehicle to Driver
func (u *DriverUsecase) AssignVehicleToDriver(ctx context.Context, driverID uint, vehicleID uint) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var driver entity.Driver
		if err := tx.First(&driver, driverID).Error; err != nil {
			return errors.New("driver not found")
		}

		var vehicle entity.Vehicle
		if err := tx.First(&vehicle, vehicleID).Error; err != nil {
			return errors.New("vehicle not found")
		}

		// Remove driver from any previously assigned vehicle
		if err := tx.Model(&entity.Vehicle{}).Where("driver_id = ?", driver.ID).Update("driver_id", nil).Error; err != nil {
			return err
		}

		// Remove vehicle from any previously assigned driver
		if err := tx.Model(&entity.Driver{}).Where("vehicle_id = ?", vehicle.ID).Update("vehicle_id", nil).Error; err != nil {
			return err
		}

		// Link vehicle and driver
		driver.VehicleID = &vehicle.ID
		if err := tx.Save(&driver).Error; err != nil {
			return err
		}

		vehicle.DriverID = &driver.ID
		vehicle.Status = "assigned"
		if err := tx.Save(&vehicle).Error; err != nil {
			return err
		}

		return nil
	})
}

// 📄 Admin: Assign Driver to Booking
func (u *DriverUsecase) AssignDriverToBooking(ctx context.Context, bookingID uint, driverID uint) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var booking entity.Booking
		if err := tx.First(&booking, bookingID).Error; err != nil {
			return errors.New("booking not found")
		}

		var driver entity.Driver
		if err := tx.First(&driver, driverID).Error; err != nil {
			return errors.New("driver not found")
		}

		booking.DriverID = &driver.ID
		if driver.VehicleID != nil {
			booking.VehicleID = driver.VehicleID
		}

		if err := tx.Save(&booking).Error; err != nil {
			return err
		}

		if driver.UserID != nil {
			notification := entity.Notification{
				UserID:    *driver.UserID,
				Title:     "New Trip Assigned",
				Message:   fmt.Sprintf("You have been assigned to Booking #%d.", booking.ID),
				IsRead:    false,
				BookingID: &booking.ID,
			}
			_ = tx.Create(&notification).Error
		}

		return nil
	})
}

// 📄 Driver: Get Dashboard Stats
func (u *DriverUsecase) GetDriverDashboard(ctx context.Context, userID uint) (*dto.DriverDashboardResponse, error) {
	driver, err := u.driverRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var bookings []entity.Booking
	if err := u.db.WithContext(ctx).
		Preload("User").
		Preload("Trip").
		Preload("Vehicle").
		Where("driver_id = ?", driver.ID).
		Order("created_at DESC").
		Find(&bookings).Error; err != nil {
		return nil, err
	}

	var total, active, completed, upcoming int
	for _, b := range bookings {
		total++
		switch b.Status {
		case "started", "ongoing":
			active++
		case "completed", "complete":
			completed++
		case "pending", "scheduled", "approved":
			upcoming++
		}
	}

	recent := make([]entity.BookingResponse, 0)
	limit := 5
	if len(bookings) < limit {
		limit = len(bookings)
	}

	for i := 0; i < limit; i++ {
		b := bookings[i]
		var userResp *entity.UserResponse
		if b.User.ID != 0 {
			userResp = &entity.UserResponse{
				ID:    b.User.ID,
				Name:  b.User.FullName,
				Email: b.User.Email,
			}
		}

		recent = append(recent, entity.BookingResponse{
			ID:            b.ID,
			Status:        b.Status,
			UserID:        b.UserID,
			User:          userResp,
			TripID:        b.TripID,
			Trip:          &b.Trip,
			SlotID:        b.SlotID,
			VehicleID:     b.VehicleID,
			BookingType:   b.BookingType,
			SeatsBooked:   b.SeatsBooked,
			BaseAmount:    b.BaseAmount,
			FinalAmount:   b.FinalAmount,
			PaymentStatus: b.PaymentStatus,
			CreatedAt:     b.CreatedAt,
			StartDate:     b.StartDate,
			EndDate:       b.EndDate,
		})
	}

	return &dto.DriverDashboardResponse{
		TotalTrips:     total,
		ActiveTrips:    active,
		CompletedTrips: completed,
		UpcomingTrips:  upcoming,
		RecentTrips:    recent,
	}, nil
}

// 📄 Driver: Get Trips
func (u *DriverUsecase) GetDriverTrips(ctx context.Context, userID uint) ([]entity.Booking, error) {
	driver, err := u.driverRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var bookings []entity.Booking
	if err := u.db.WithContext(ctx).
		Preload("User").
		Preload("Trip").
		Preload("Vehicle").
		Where("driver_id = ?", driver.ID).
		Order("created_at DESC").
		Find(&bookings).Error; err != nil {
		return nil, err
	}

	return bookings, nil
}

// 📄 Driver: Get Trip Details
func (u *DriverUsecase) GetDriverTripByID(ctx context.Context, userID uint, bookingID uint) (*entity.Booking, error) {
	driver, err := u.driverRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var booking entity.Booking
	if err := u.db.WithContext(ctx).
		Preload("User").
		Preload("Trip").
		Preload("Vehicle").
		Where("id = ? AND driver_id = ?", bookingID, driver.ID).
		First(&booking).Error; err != nil {
		return nil, errors.New("trip not found or not assigned to you")
	}

	return &booking, nil
}

// 📄 Driver: Update Trip Status
func (u *DriverUsecase) UpdateTripStatus(ctx context.Context, userID uint, bookingID uint, newStatus string) error {
	driver, err := u.driverRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var booking entity.Booking
		if err := tx.Where("id = ? AND driver_id = ?", bookingID, driver.ID).First(&booking).Error; err != nil {
			return errors.New("booking not found or access denied")
		}

		booking.Status = newStatus
		if err := tx.Save(&booking).Error; err != nil {
			return err
		}

		notification := entity.Notification{
			UserID:    booking.UserID,
			Title:     "Trip Status Update",
			Message:   fmt.Sprintf("Your trip status has been updated to: %s", newStatus),
			IsRead:    false,
			BookingID: &booking.ID,
		}
		_ = tx.Create(&notification).Error

		return nil
	})
}

// 📄 Driver: Get Vehicle Details
func (u *DriverUsecase) GetDriverVehicle(ctx context.Context, userID uint) (*entity.Vehicle, error) {
	driver, err := u.driverRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if driver.VehicleID == nil {
		return nil, errors.New("no vehicle assigned to you")
	}

	var vehicle entity.Vehicle
	if err := u.db.WithContext(ctx).First(&vehicle, *driver.VehicleID).Error; err != nil {
		return nil, errors.New("assigned vehicle details not found")
	}

	return &vehicle, nil
}

// 📄 Driver: Get Profile
func (u *DriverUsecase) GetDriverProfile(ctx context.Context, userID uint) (*entity.Driver, error) {
	return u.driverRepo.GetByUserID(ctx, userID)
}

// 📄 Driver: Update Profile
func (u *DriverUsecase) UpdateDriverProfile(ctx context.Context, userID uint, req dto.DriverProfileUpdateRequest) error {
	driver, err := u.driverRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if req.Phone != "" {
			driver.Phone = req.Phone
		}
		if req.Address != "" {
			driver.Address = req.Address
		}

		if err := tx.Save(driver).Error; err != nil {
			return err
		}

		if req.OldPassword != "" && req.NewPassword != "" {
			var user entity.User
			if err := tx.First(&user, userID).Error; err != nil {
				return err
			}

			if !Checkpassword(req.OldPassword, user.HashPassword) {
				return errors.New("incorrect old password")
			}

			newHash, err := HashPassword(req.NewPassword)
			if err != nil {
				return err
			}

			user.HashPassword = newHash
			if err := tx.Save(&user).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// 📄 Driver: Update Live Location Tracking
func (u *DriverUsecase) UpdateDriverTracking(ctx context.Context, userID uint, lat float64, lng float64) error {
	driver, err := u.driverRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	var booking entity.Booking
	err = u.db.WithContext(ctx).
		Where("driver_id = ? AND status IN ?", driver.ID, []string{"started", "ongoing"}).
		First(&booking).Error

	if err != nil {
		return errors.New("no active started/ongoing trip to send tracking coordinates")
	}

	tracking := entity.Tracking{
		BookingID: booking.ID,
		VehicleID: driver.VehicleID,
		DriverID:  &driver.ID,
		Type:      "driver",
		Latitude:  lat,
		Longitude: lng,
	}

	return u.trackingRepo.Create(ctx, &tracking)
}
