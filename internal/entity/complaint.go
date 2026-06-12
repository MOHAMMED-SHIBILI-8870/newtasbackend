package entity

import "time"

type Complaint struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	BookingID   uint      `gorm:"not null;index" json:"booking_id"`
	Title       string    `gorm:"size:150;not null" json:"title"`
	Description string    `gorm:"type:text;not null" json:"description"`
	Status      string    `gorm:"size:30;not null;default:'pending';index" json:"status"`
	AdminID     *uint     `gorm:"index" json:"admin_id,omitempty"`
	AdminNotes  string    `gorm:"type:text" json:"admin_notes,omitempty"`
	User        User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;" json:"-"`
	Booking     Booking   `gorm:"foreignKey:BookingID;constraint:OnDelete:CASCADE;" json:"-"`
	Admin       *User     `gorm:"foreignKey:AdminID;constraint:OnDelete:SET NULL;" json:"-"`
}
