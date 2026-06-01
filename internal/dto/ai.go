package dto

import "time"

type AITripRequestInput struct {
	From          string `json:"from"`
	To            string `json:"to"`
	Days          int    `json:"days"`
	TripType      string `json:"trip_type"`
	BudgetLevel   string `json:"budget_level"`
	Members       int    `json:"members"`
	Children      int    `json:"children"`
	HotelType     string `json:"hotel_type"`
	Transport     string `json:"transport"`
	Prompt        string `json:"prompt"`
	GeneratedPlan string `json:"generated_plan"`
}

type AITripReviewRequest struct {
	AdminNote string `json:"admin_note"`
}

type AITripRequestUser struct {
	ID       uint   `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type AITripRequestResponse struct {
	ID            uint              `json:"id"`
	User          AITripRequestUser `json:"user"`
	From          string            `json:"from"`
	To            string            `json:"to"`
	Days          int               `json:"days"`
	TripType      string            `json:"trip_type"`
	BudgetLevel   string            `json:"budget_level"`
	Members       int               `json:"members"`
	Children      int               `json:"children"`
	HotelType     string            `json:"hotel_type"`
	Transport     string            `json:"transport"`
	Prompt        string            `json:"prompt"`
	GeneratedPlan string            `json:"generated_plan"`
	Status        string            `json:"status"`
	AdminNote     string            `json:"admin_note,omitempty"`
	TripID        *uint             `json:"trip_id,omitempty"`
	ReviewedByID  *uint             `json:"reviewed_by_id,omitempty"`
	ReviewedAt    *time.Time        `json:"reviewed_at,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}
