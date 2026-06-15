package repository

import (
	"backend/internal/entity"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostgresChatRepository struct {
	db *gorm.DB
}

func NewPostgresChatRepository(db *gorm.DB) entity.ChatRepository {
	return &PostgresChatRepository{db: db}
}

func (r *PostgresChatRepository) GetRoom(ctx context.Context, userID, guideID string) (*entity.ChatRoom, error) {
	var room entity.ChatRoom
	err := r.db.WithContext(ctx).
		Where(`
			(user_id = ? AND guide_id = ?) OR (guide_id = ? AND user_id = ?) OR
			(user_id = ? AND support_agent_id = ?) OR (support_agent_id = ? AND user_id = ?) OR
			(admin_id = ? AND support_agent_id = ?) OR (support_agent_id = ? AND admin_id = ?)
		`, userID, guideID, userID, guideID, userID, guideID, userID, guideID, userID, guideID, userID, guideID).
		First(&room).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("chat room not found")
		}
		return nil, err
	}
	return &room, nil
}

func (r *PostgresChatRepository) CreateRoom(ctx context.Context, room *entity.ChatRoom) error {
	if room.ID == "" {
		room.ID = uuid.New().String()
	}
	if room.CreatedAt.IsZero() {
		room.CreatedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(room).Error
}

func (r *PostgresChatRepository) SaveMessage(ctx context.Context, msg *entity.Message) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *PostgresChatRepository) GetMessagesByRoom(ctx context.Context, roomID string, limit int) ([]entity.Message, error) {
	var msgs []entity.Message
	err := r.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("created_at ASC"). // Typically chat messages are ordered oldest to newest for UI
		Limit(limit).
		Find(&msgs).Error

	return msgs, err
}

func (r *PostgresChatRepository) ValidatePermission(ctx context.Context, userID, targetID string) (bool, error) {
	if userID == "" || targetID == "" {
		return false, nil
	}
	return true, nil
}

func (r *PostgresChatRepository) MarkMessagesAsRead(ctx context.Context, roomID string, readerID string) error {
	// Mark messages in this room as read if the reader is the receiver of the message
	return r.db.WithContext(ctx).
		Model(&entity.Message{}).
		Where("room_id = ? AND receiver_id = ? AND is_read = ?", roomID, readerID, false).
		Update("is_read", true).Error
}

func (r *PostgresChatRepository) GetRoomsForUser(ctx context.Context, userID string) ([]entity.ChatRoom, error) {
	var rooms []entity.ChatRoom
	err := r.db.WithContext(ctx).
		Where(`user_id = ? OR guide_id = ? OR support_agent_id = ? OR admin_id = ?`, userID, userID, userID, userID).
		Find(&rooms).Error
	return rooms, err
}
