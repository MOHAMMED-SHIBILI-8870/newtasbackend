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

// Compile-time interface check
var _ entity.ChatRepository = (*MemoryChatRepository)(nil)

func NewMemoryChatRepository() entity.ChatRepository {
	return &MemoryChatRepository{
		rooms:    make(map[string]*entity.ChatRoom),
		messages: make(map[string][]entity.Message),
	}
}

func (r *MemoryChatRepository) GetRoom(
	ctx context.Context,
	userID,
	guideID string,
) (*entity.ChatRoom, error) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, room := range r.rooms {
		if (room.UserID == userID && room.GuideID == guideID) ||
			(room.UserID == guideID && room.GuideID == userID) {
			return room, nil
		}
	}

	return nil, errors.New("chat room not found")
}

func (r *MemoryChatRepository) CreateRoom(
	ctx context.Context,
	room *entity.ChatRoom,
) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	if room.ID == "" {
		room.ID = uuid.New().String()
	}

	room.CreatedAt = time.Now()

	r.rooms[room.ID] = room

	return nil
}

func (r *MemoryChatRepository) SaveMessage(
	ctx context.Context,
	msg *entity.Message,
) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	msg.ID = uuid.New().String()
	msg.CreatedAt = time.Now()

	r.messages[msg.RoomID] = append(
		r.messages[msg.RoomID],
		*msg,
	)

	return nil
}

func (r *MemoryChatRepository) GetMessagesByRoom(
	ctx context.Context,
	roomID string,
	limit int,
) ([]entity.Message, error) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	msgs, exists := r.messages[roomID]
	if !exists {
		return []entity.Message{}, nil
	}

	if limit > 0 && len(msgs) > limit {
		msgs = msgs[len(msgs)-limit:]
	}

	return msgs, nil
}

func (r *MemoryChatRepository) ValidatePermission(
	ctx context.Context,
	userID,
	guideID string,
) (bool, error) {

	if userID == "" || guideID == "" {
		return false, nil
	}

	// TODO:
	// Production version should verify
	// booking relationship between user and guide.

	return true, nil
}

func (r *MemoryChatRepository) MarkMessagesAsRead(
	ctx context.Context,
	roomID string,
	readerID string,
) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	msgs, exists := r.messages[roomID]
	if !exists {
		return nil
	}

	for i := range msgs {
		if msgs[i].SenderID != readerID {
			msgs[i].IsRead = true
		}
	}

	r.messages[roomID] = msgs

	return nil
}

func (r *MemoryChatRepository) GetRoomsForUser(
	ctx context.Context,
	userID string,
) ([]entity.ChatRoom, error) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	var rooms []entity.ChatRoom

	for _, room := range r.rooms {

		if room.UserID == userID ||
			room.GuideID == userID ||
			room.SupportAgentID == userID ||
			room.AdminID == userID {

			rooms = append(rooms, *room)
		}
	}

	return rooms, nil
}