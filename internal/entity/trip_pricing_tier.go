package entity

type TripPricingTier struct {
	ID      uint    `gorm:"primaryKey" json:"id"`
	TripID  uint    `gorm:"not null;index;constraint:OnDelete:CASCADE;" json:"trip_id"`
	Members int     `gorm:"not null" json:"members"`
	Price   float64 `gorm:"type:decimal(10,2);not null" json:"price"`
}
