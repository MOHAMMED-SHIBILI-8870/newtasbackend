package dto

import "time"

type NotificationResponse struct {
	ID             uint      `json:"id"`
	Type           string    `json:"type"`
	Title          string    `json:"title"`
	Message        string    `json:"message"`
	UserID         uint      `json:"user_id"`
	BookingID      *uint     `json:"booking_id,omitempty"`
	AITripRequestID *uint    `json:"ai_trip_request_id,omitempty"`
	IsRead         bool      `json:"is_read"`
	CreatedAt      time.Time `json:"created_at"`
}

type MarkNotificationReadRequest struct {
	NotificationID uint `json:"notification_id"`
}
