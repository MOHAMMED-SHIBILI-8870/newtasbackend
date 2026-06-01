package dto

import "time"

type ComplaintRequest struct {
	BookingID   uint   `json:"booking_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type ComplaintStatusRequest struct {
	Status string `json:"status"`
}

type ComplaintResponse struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	BookingID   uint      `json:"booking_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
