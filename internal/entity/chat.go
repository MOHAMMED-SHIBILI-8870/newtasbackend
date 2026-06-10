package entity

import (
	"context"
	"time"
)

// Message represents a single chat message
type Message struct {
	ID         string    `json:"id"`
	RoomID     string    `json:"room_id"`
	SenderID   string    `json:"sender_id"`
	ReceiverID string    `json:"receiver_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

// ChatRoom represents a 1-to-1 conversation space between a User and a Guide
type ChatRoom struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	GuideID   string    `json:"guide_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Repositories interfaces (Inversion of Control)
type ChatRepository interface {
	GetRoom(ctx context.Context, userID, guideID string) (*ChatRoom, error)
	CreateRoom(ctx context.Context, room *ChatRoom) error
	SaveMessage(ctx context.Context, msg *Message) error
	GetMessagesByRoom(ctx context.Context, roomID string, limit int) ([]Message, error)
	ValidatePermission(ctx context.Context, userID, guideID string) (bool, error)
}

// Usecase interfaces
type ChatUsecase interface {
	SendMessage(ctx context.Context, senderID, receiverID, content string) (*Message, error)
	GetMessages(ctx context.Context, userID, guideID string) ([]Message, error)
}