package entity

import "time"

type Permission struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Key         string    `gorm:"size:100;uniqueIndex;not null" json:"key"`
	Name        string    `gorm:"size:150;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
