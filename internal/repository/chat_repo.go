package repository

import (
	"backend/internal/entity"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

type MemoryChatRepository struct {
	mu       sync.RWMutex
	rooms    map[string]*entity.ChatRoom
	messages map[string][]entity.Message
}

func NewMemoryChatRepository() entity.ChatRepository {
	return &MemoryChatRepository{
		rooms:    make(map[string]*entity.ChatRoom),
		messages: make(map[string][]entity.Message),
	}
}

func (r *MemoryChatRepository) GetRoom(ctx context.Context, userID, guideID string) (*entity.ChatRoom, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, room := range r.rooms {
		if (room.UserID == userID && room.GuideID == guideID) || (room.UserID == guideID && room.GuideID == userID) {
			return room, nil
		}
	}
	return nil, errors.New("chat room not found")
}

func (r *MemoryChatRepository) CreateRoom(ctx context.Context, room *entity.ChatRoom) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if room.ID == "" {
		room.ID = uuid.New().String()
	}
	room.CreatedAt = time.Now()
	r.rooms[room.ID] = room
	return nil
}

func (r *MemoryChatRepository) SaveMessage(ctx context.Context, msg *entity.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	msg.ID = uuid.New().String()
	msg.CreatedAt = time.Now()
	r.messages[msg.RoomID] = append(r.messages[msg.RoomID], *msg)
	return nil
}

func (r *MemoryChatRepository) GetMessagesByRoom(ctx context.Context, roomID string, limit int) ([]entity.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	msgs, exists := r.messages[roomID]
	if !exists {
		return []entity.Message{}, nil
	}
	return msgs, nil
}

// ValidatePermission acts as the User-Guide mapping validator rule
func (r *MemoryChatRepository) ValidatePermission(ctx context.Context, userID, guideID string) (bool, error) {
	// Business Rule: Explicit mapping verification. 
	// For production, check a `bookings` or `contracts` table.
	if userID == "" || guideID == "" {
		return false, nil
	}
	return true, nil // Mocked to true for demonstration
}

func (r *MemoryChatRepository) MarkMessagesAsRead(ctx context.Context, roomID string, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	messages, exists := r.messages[roomID]
	if !exists {
		return nil // No messages to update
	}

	// Loop through and mark messages as read if they belong to the other user
	for i := range messages {
		// Assuming your entity.Message has a SenderID (or UserID) and IsRead field.
		// We only mark messages as read if the current user is NOT the sender.
		if messages[i].SenderID != userID { 
			messages[i].IsRead = true
		}
	}

	// Save the modified slice back to the map
	r.messages[roomID] = messages
	return nil
}