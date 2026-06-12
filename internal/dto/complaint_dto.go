package dto

import "time"

type ComplaintRequest struct {
	BookingID   uint   `json:"booking_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type ComplaintStatusRequest struct {
	Status     string `json:"status"`
	AdminID    *uint  `json:"admin_id,omitempty"`
	AdminNotes string `json:"admin_notes,omitempty"`
}

type ComplaintResponse struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	BookingID   uint      `json:"booking_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	AdminID     *uint     `json:"admin_id,omitempty"`
	AdminNotes  string    `json:"admin_notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
