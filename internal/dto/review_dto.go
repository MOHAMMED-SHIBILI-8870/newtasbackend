package dto

import "time"

type ReviewRequest struct {
	TripID  uint   `json:"trip_id"`
	GuideID *uint  `json:"guide_id,omitempty"`
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

type ReviewResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	TripID    uint      `json:"trip_id"`
	GuideID   *uint     `json:"guide_id,omitempty"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User *ReviewUserResponse `json:"user,omitempty"`
	Trip *ReviewTripResponse `json:"trip,omitempty"`
}

type ReviewUserResponse struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

type ReviewTripResponse struct {
	From string `json:"from"`
	To   string `json:"to"`
}


type ReviewSummaryResponse struct {
	TripID        uint    `json:"trip_id"`
	AverageRating float64 `json:"average_rating"`
	ReviewCount   int64   `json:"review_count"`
}
