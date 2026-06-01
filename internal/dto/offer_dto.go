package dto

import "time"

type OfferRequest struct {
	Code            string    `json:"code"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	DiscountPercent float64   `json:"discount_percent"`
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
	ExpiryDate      time.Time `json:"expiry_date"`
	Active          bool      `json:"active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
