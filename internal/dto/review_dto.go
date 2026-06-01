package dto

import "time"

type ReviewRequest struct {
	TripID  uint   `json:"trip_id"`
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

type ReviewResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	TripID    uint      `json:"trip_id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ReviewSummaryResponse struct {
	TripID        uint    `json:"trip_id"`
	AverageRating float64 `json:"average_rating"`
	ReviewCount   int64   `json:"review_count"`
}
