package entity

import "time"

type UpdateTripInput struct {
	From         *string    `json:"from"`
	To           *string    `json:"to"`
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	Duration                 *int       `json:"duration"`
	MinDuration              *int       `json:"min_duration"`
	TripType                 *string    `json:"trip_type"`
	BudgetLevel              *string    `json:"budget_level"`
	Price                    *float64   `json:"price"`
	Members                  *int       `json:"members"`
	Children                 *int       `json:"children"`
	ConcurrentSlots          *int       `json:"concurrent_slots"`
	GroupDiscountThreshold   *int       `json:"group_discount_threshold"`
	GroupDiscountPercent     *float64   `json:"group_discount_percent"`
	HotelType    *string    `json:"hotel_type"`
	Transport    *string    `json:"transport"`
	ItineraryRaw *string    `json:"itinerary_raw"`
	ImageURL     *string    `json:"image_url"`
	Status       *string    `json:"status"`
	Plans        []TripPlan        `json:"plans"`
	PricingTiers []TripPricingTier `json:"pricing_tiers"`
}
