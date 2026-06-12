package dto

import "time"

type OfferRequest struct {
	Code            string    `json:"code"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	DiscountPercent float64   `json:"discount_percent"`
	DiscountType    string    `json:"discount_type"`
	FixedDiscount   float64   `json:"fixed_discount"`
	MaxUsage        int       `json:"max_usage"`
	TripID          *uint     `json:"trip_id,omitempty"`
	ExpiryDate      time.Time `json:"expiry_date"`
	Active          bool      `json:"active"`
}

type ApplyCouponRequest struct {
	Code string `json:"code"`
}

type OfferResponse struct {
	ID              uint      `json:"id"`
	Code            string    `json:"code"`
	Title           string    `json:"title"`
	Description     string    `json:"description,omitempty"`
	DiscountPercent float64   `json:"discount_percent"`
	DiscountType    string    `json:"discount_type"`
	FixedDiscount   float64   `json:"fixed_discount"`
	MaxUsage        int       `json:"max_usage"`
	CurrentUsage    int       `json:"current_usage"`
	TripID          *uint     `json:"trip_id,omitempty"`
	ExpiryDate      time.Time `json:"expiry_date"`
	Active          bool      `json:"active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
