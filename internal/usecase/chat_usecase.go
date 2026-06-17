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

	if receiverID == "support" {
		if u.userRepo != nil {
			supportUser, err := u.userRepo.GetByEmail(ctx, "support@gmail.com")
			if err == nil && supportUser != nil {
				receiverID = strconv.Itoa(int(supportUser.ID))
			}
		}
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
	if guideID == "support" {
		supportAgentID, err := u.GetSupportAgentID(ctx)
		if err == nil {
			guideID = supportAgentID
		}
	}

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
	if guideID == "support" {
		supportAgentID, err := u.GetSupportAgentID(ctx)
		if err == nil {
			guideID = supportAgentID
		}
	}

	room, err := u.repo.GetRoom(ctx, userID, guideID)
	if err != nil {
		return err
	}
	return u.repo.MarkMessagesAsRead(ctx, room.ID, readerID)
}

func (u *chatUsecase) GetContacts(ctx context.Context, userID string) ([]entity.ContactInfo, error) {
	rooms, err := u.repo.GetRoomsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	contactSet := make(map[string]string)
	for _, room := range rooms {
		if room.UserID != "" && room.UserID != userID {
			contactSet[room.UserID] = "user"
		}
		if room.GuideID != "" && room.GuideID != userID {
			contactSet[room.GuideID] = "guide"
		}
		if room.SupportAgentID != "" && room.SupportAgentID != userID {
			contactSet[room.SupportAgentID] = "support"
		}
		if room.AdminID != "" && room.AdminID != userID {
			contactSet[room.AdminID] = "admin"
		}
	}

	contacts := make([]entity.ContactInfo, 0, len(contactSet))
	for contactID, contactType := range contactSet {
		name := "User ID: " + contactID
		if u.userRepo != nil {
			idUint, err := strconv.Atoi(contactID)
			if err == nil {
				user, _ := u.userRepo.GetByID(ctx, uint(idUint))
				if user != nil && user.FullName != "" {
					name = user.FullName
				}
			}
		}
		
		if contactType == "support" {
			name = "Customer Support"
		}

		contacts = append(contacts, entity.ContactInfo{
			ID:   contactID,
			Name: name,
			Type: contactType,
		})
	}

	return contacts, nil
}

func (u *chatUsecase) GetSupportAgentID(ctx context.Context) (string, error) {
	if u.userRepo != nil {
		supportUser, err := u.userRepo.GetByEmail(ctx, "support@gmail.com")
		if err == nil && supportUser != nil {
			return strconv.Itoa(int(supportUser.ID)), nil
		}
	}
	return "4", nil // Fallback to seeded ID 4
}
