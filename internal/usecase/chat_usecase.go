package usecase

import (
	"backend/internal/entity"
	"backend/internal/repository"
	"context"
	"errors"
	"strconv"
)

type chatUsecase struct {
	repo entity.ChatRepository
	userRepo repository.UserRepository
}

func NewChatUsecase(r entity.ChatRepository, userRepo repository.UserRepository) entity.ChatUsecase {
	return &chatUsecase{repo: r, userRepo: userRepo}
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

	// Fetch roles to set appropriate room type
	senderIDUint, _ := strconv.Atoi(senderID)
	receiverIDUint, _ := strconv.Atoi(receiverID)

	var senderRole, receiverRole string
	if u.userRepo != nil {
		senderUser, _ := u.userRepo.GetByID(ctx, uint(senderIDUint))
		if senderUser != nil {
			senderRole = NormalizeRole(senderUser.Role)
		}
		receiverUser, _ := u.userRepo.GetByID(ctx, uint(receiverIDUint))
		if receiverUser != nil {
			receiverRole = NormalizeRole(receiverUser.Role)
		}
	}

	// 2. Fetch or Create Chat Room between the pair
	room, err := u.repo.GetRoom(ctx, senderID, receiverID)
	if err != nil {
		room = &entity.ChatRoom{}
		
		if (senderRole == "user" && receiverRole == "guide") || (senderRole == "guide" && receiverRole == "user") {
			room.RoomType = "user_guide"
			if senderRole == "user" {
				room.UserID = senderID
				room.GuideID = receiverID
			} else {
				room.UserID = receiverID
				room.GuideID = senderID
			}
		} else if (senderRole == "user" && receiverRole == "supportagent") || (senderRole == "supportagent" && receiverRole == "user") {
			room.RoomType = "user_support"
			if senderRole == "user" {
				room.UserID = senderID
				room.SupportAgentID = receiverID
			} else {
				room.UserID = receiverID
				room.SupportAgentID = senderID
			}
		} else if (senderRole == "admin" && receiverRole == "supportagent") || (senderRole == "supportagent" && receiverRole == "admin") {
			room.RoomType = "admin_support"
			if senderRole == "admin" {
				room.AdminID = senderID
				room.SupportAgentID = receiverID
			} else {
				room.AdminID = receiverID
				room.SupportAgentID = senderID
			}
		} else {
			room.RoomType = "user_guide"
			room.UserID = senderID
			room.GuideID = receiverID
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
		IsRead:     false, // explicitly false initially
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

	// When history is fetched, we could optionally mark messages as read here,
	// but usually that is a separate endpoint or done via WebSocket. 
	// For now, we just return the history.
	return u.repo.GetMessagesByRoom(ctx, room.ID, 100)
}

func (u *chatUsecase) MarkMessagesAsRead(ctx context.Context, userID, guideID, readerID string) error {
	room, err := u.repo.GetRoom(ctx, userID, guideID)
	if err != nil {
		return err
	}
	return u.repo.MarkMessagesAsRead(ctx, room.ID, readerID)
}
