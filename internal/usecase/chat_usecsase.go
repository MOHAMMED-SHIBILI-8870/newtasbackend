package usecase

import (
	"backend/internal/entity"
	"context"
	"errors"
)

type chatUsecase struct {
	repo entity.ChatRepository
}

func NewChatUsecase(r entity.ChatRepository) entity.ChatUsecase {
	return &chatUsecase{repo: r}
}

func (u *chatUsecase) SendMessage(ctx context.Context, senderID, receiverID, content string) (*entity.Message, error) {
	if content == "" {
		return nil, errors.New("message content cannot be empty")
	}

	// 1. Validate explicit 1-to-1 access permission rules
	hasAccess, err := u.repo.ValidatePermission(ctx, senderID, receiverID)
	if err != nil || !hasAccess {
		return nil, errors.New("unauthorized: user-guide relationship invalid")
	}

	// 2. Fetch or Create Chat Room between the pair
	room, err := u.repo.GetRoom(ctx, senderID, receiverID)
	if err != nil {
		// Room doesn't exist, create it (assigning sender/receiver properly to User/Guide fields)
		room = &entity.ChatRoom{
			UserID:  senderID,
			GuideID: receiverID,
		}
		if err := u.repo.CreateRoom(ctx, room); err != nil {
			return nil, err
		}
	}

	// 3. Build and persist message
	msg := &entity.Message{
		RoomID:     room.ID,
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
	}

	if err := u.repo.SaveMessage(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (u *chatUsecase) GetMessages(ctx context.Context, userID, guideID string) ([]entity.Message, error) {
	// Validate authorization rule
	hasAccess, err := u.repo.ValidatePermission(ctx, userID, guideID)
	if err != nil || !hasAccess {
		return nil, errors.New("unauthorized to view this chat history")
	}

	room, err := u.repo.GetRoom(ctx, userID, guideID)
	if err != nil {
		return []entity.Message{}, nil // No history yet
	}

	return u.repo.GetMessagesByRoom(ctx, room.ID, 100)
}