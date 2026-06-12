package entity

import "time"

type SupportRequest struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index" json:"user_id"`
	AgentID     *uint     `gorm:"index" json:"agent_id,omitempty"` // The assigned support agent
	Status      string    `gorm:"type:varchar(50);default:'open'" json:"status"` // open, in_progress, resolved
	Subject     string    `gorm:"type:varchar(255);not null" json:"subject"`
	Description string    `gorm:"type:text;not null" json:"description"`
	ChatRoomID  string    `gorm:"type:uuid" json:"chat_room_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	User  User  `gorm:"foreignKey:UserID" json:"-"`
	Agent *User `gorm:"foreignKey:AgentID" json:"-"`
}
