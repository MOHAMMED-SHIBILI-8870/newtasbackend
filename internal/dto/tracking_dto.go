package dto

import "time"

type TrackingUpdateRequest struct {
	BookingID uint    `json:"booking_id"`
	VehicleID uint    `json:"vehicle_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type TrackingResponse struct {
	ID        uint      `json:"id"`
	BookingID uint      `json:"booking_id"`
	VehicleID uint      `json:"vehicle_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
