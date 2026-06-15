package entity

import (
	"context"
	"time"
)

// Message represents a single chat message
type Message struct {
	ID         string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	RoomID     string    `gorm:"type:uuid;not null;index" json:"room_id"`
	SenderID   string    `gorm:"type:varchar(255);not null;index" json:"sender_id"`
	ReceiverID string    `gorm:"type:varchar(255);not null;index" json:"receiver_id"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	IsRead     bool      `gorm:"default:false" json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
}

// ChatRoom represents a 1-to-1 conversation space between a User and a Guide
type ChatRoom struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID         string    `gorm:"type:varchar(255);index" json:"user_id,omitempty"`
	GuideID        string    `gorm:"type:varchar(255);index" json:"guide_id,omitempty"`
	SupportAgentID string    `gorm:"type:varchar(255);index" json:"support_agent_id,omitempty"`
	AdminID        string    `gorm:"type:varchar(255);index" json:"admin_id,omitempty"`
	RoomType       string    `gorm:"type:varchar(50);default:'user_guide'" json:"room_type"`
	CreatedAt      time.Time `json:"created_at"`

	Messages  []Message `gorm:"foreignKey:RoomID;constraint:OnDelete:CASCADE" json:"messages,omitempty"`
}

// Repositories interfaces (Inversion of Control)
type ChatRepository interface {
	GetRoom(ctx context.Context, userID, guideID string) (*ChatRoom, error)
	CreateRoom(ctx context.Context, room *ChatRoom) error
	SaveMessage(ctx context.Context, msg *Message) error
	GetMessagesByRoom(ctx context.Context, roomID string, limit int) ([]Message, error)
	ValidatePermission(ctx context.Context, userID, guideID string) (bool, error)
	MarkMessagesAsRead(ctx context.Context, roomID string, readerID string) error
	GetRoomsForUser(ctx context.Context, userID string) ([]ChatRoom, error)
}

// Usecase interfaces
type ChatUsecase interface {
	SendMessage(ctx context.Context, senderID, receiverID, content string) (*Message, error)
	GetMessages(ctx context.Context, userID, guideID string) ([]Message, error)
	MarkMessagesAsRead(ctx context.Context, userID, guideID, readerID string) error
	GetContacts(ctx context.Context, userID string) ([]string, error)
	GetSupportAgentID(ctx context.Context) (string, error)
}