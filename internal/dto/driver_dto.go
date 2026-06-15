package dto

import (
	"backend/internal/entity"
	"time"
)

type DriverCreateRequest struct {
	Name          string    `json:"name" validate:"required"`
	Email         string    `json:"email" validate:"required,email"`
	Password      string    `json:"password" validate:"required,min=6"`
	Phone         string    `json:"phone" validate:"required"`
	Address       string    `json:"address"`
	LicenseNumber string    `json:"license_number" validate:"required"`
	LicenseExpiry time.Time `json:"license_expiry" validate:"required"`
}

type DriverUpdateRequest struct {
	Name          string     `json:"name"`
	Email         string     `json:"email"`
	Phone         string     `json:"phone"`
	Address       string     `json:"address"`
	LicenseNumber string     `json:"license_number"`
	LicenseExpiry *time.Time `json:"license_expiry"`
	Status        string     `json:"status"`
	VehicleID     *uint      `json:"vehicle_id"`
}

type DriverProfileUpdateRequest struct {
	Phone       string `json:"phone"`
	Address     string `json:"address"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type DriverResponse struct {
	ID            uint             `json:"id"`
	UserID        *uint            `json:"user_id,omitempty"`
	Name          string           `json:"name"`
	Email         string           `json:"email"`
	Phone         string           `json:"phone"`
	Address       string           `json:"address"`
	LicenseNumber string           `json:"license_number"`
	LicenseExpiry time.Time        `json:"license_expiry"`
	Status        string           `json:"status"`
	VehicleID     *uint            `json:"vehicle_id,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	Vehicle       *VehicleResponse `json:"vehicle,omitempty"`
}

type DriverDashboardResponse struct {
	TotalTrips     int               `json:"total_trips"`
	ActiveTrips    int               `json:"active_trips"`
	CompletedTrips int               `json:"completed_trips"`
	UpcomingTrips  int               `json:"upcoming_trips"`
	RecentTrips    []entity.BookingResponse `json:"recent_trips"`
}

type TripStatusUpdateRequest struct {
	Status string `json:"status" validate:"required"`
}

type DriverTrackingUpdateRequest struct {
	Latitude  float64 `json:"latitude" validate:"required"`
	Longitude float64 `json:"longitude" validate:"required"`
}
