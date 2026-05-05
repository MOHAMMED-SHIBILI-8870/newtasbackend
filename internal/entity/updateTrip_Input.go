package entity

type UpdateTripInput struct {
	Destination *string  `json:"destination"`
	Budget      *float64 `json:"budget"`
	Duration    *int     `json:"duration"`
	Description *string  `json:"description"`
}