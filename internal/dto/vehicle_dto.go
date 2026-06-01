package dto

import "time"

type VehicleRequest struct {
	AgencyID       uint    `json:"agency_id"`
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	TotalSeats     int     `json:"total_seats"`
	AvailableSeats int     `json:"available_seats"`
	PricePerPerson float64 `json:"price_per_person"`
	Status         string  `json:"status"`
	TripID         *uint   `json:"trip_id,omitempty"`
}

type AssignVehicleRequest struct {
	TripID uint `json:"trip_id"`
}

type VehicleResponse struct {
	ID             uint      `json:"id"`
	AgencyID       uint      `json:"agency_id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	TotalSeats     int       `json:"total_seats"`
	AvailableSeats int       `json:"available_seats"`
	PricePerPerson float64   `json:"price_per_person"`
	Status         string    `json:"status"`
	TripID         *uint     `json:"trip_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
