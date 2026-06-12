package handler

import (
	"backend/internal/entity"
	"backend/internal/middleware"
	"context"
	"fmt"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// Hub maintains the set of active clients and coordinates real-time delivery
type Hub struct {
	clients    map[string]*websocket.Conn // key: UserID/GuideID
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.RWMutex
}

type ChatHandler struct {
	usecase entity.ChatUsecase
	hub     *Hub
}

// NewChatHandler instantiates the handler structure with dependencies injected.
// Routing concerns have been cleanly migrated out to the routes package.
func NewChatHandler(uc entity.ChatUsecase) *ChatHandler {
	return &ChatHandler{
		usecase: uc,
		hub: &Hub{
			clients: make(map[string]*websocket.Conn),
		},
	}
}

type SendMsgRequest struct {
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
	Content    string `json:"content"`
}

// REST Fallback: Send Message
func (h *ChatHandler) RESTSendMessage(c *fiber.Ctx) error {
	var req SendMsgRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	msg, err := h.usecase.SendMessage(context.Background(), req.SenderID, req.ReceiverID, req.Content)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	// If the recipient is currently online via WebSocket, dispatch instantly
	h.hub.mu.RLock()
	if conn, online := h.hub.clients[req.ReceiverID]; online {
		_ = conn.WriteJSON(msg)
	}
	h.hub.mu.RUnlock()

	return c.Status(fiber.StatusCreated).JSON(msg)
}

// REST Fallback: Fetch Chat History
func (h *ChatHandler) RESTFetchHistory(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	guideID := c.Query("guide_id")

	if userID == "" || guideID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id and guide_id are required"})
	}

	// Security: verify the authenticated caller is one of the chat participants
	callerID := middleware.GetAuthUserID(c)
	callerRole := middleware.GetAuthRole(c)
	callerIDStr := fmt.Sprintf("%d", callerID)

	if callerRole != "admin" && callerIDStr != userID && callerIDStr != guideID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "access denied"})
	}

	// Mark unread messages as read
	_ = h.usecase.MarkMessagesAsRead(context.Background(), userID, guideID, callerIDStr)

	messages, err := h.usecase.GetMessages(context.Background(), userID, guideID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	h.hub.mu.RLock()
	partnerID := guideID
	if callerIDStr == guideID {
		partnerID = userID
	}
	_, isOnline := h.hub.clients[partnerID]
	h.hub.mu.RUnlock()

	return c.JSON(fiber.Map{
		"messages": messages,
		"is_online": isOnline,
	})
}

// WebSocket connection handler
func (h *ChatHandler) WebSocketHandler(c *websocket.Conn) {
	userID := c.Query("user_id")
	if userID == "" {
		_ = c.WriteJSON(fiber.Map{"error": "unauthorized connection target missing query param: user_id"})
		c.Close()
		return
	}

	// Security: verify the authenticated user matches the requested user_id.
	// The auth middleware already ran before the WebSocket upgrade, so the
	// authenticated user ID is available via c.Locals().
	authID, _ := c.Locals("auth_user_id").(uint)
	if authID == 0 {
		// Try other possible types that may have been stored
		switch v := c.Locals("auth_user_id").(type) {
		case int:
			authID = uint(v)
		case int64:
			authID = uint(v)
		case float64:
			authID = uint(v)
		}
	}

	expectedID := fmt.Sprintf("%d", authID)
	if authID == 0 || expectedID != userID {
		_ = c.WriteJSON(fiber.Map{"error": "access denied: user_id does not match authenticated identity"})
		c.Close()
		return
	}

	// Register Client
	h.hub.mu.Lock()
	h.hub.clients[userID] = c
	h.hub.mu.Unlock()

	defer func() {
		h.hub.mu.Lock()
		delete(h.hub.clients, userID)
		h.hub.mu.Unlock()
		c.Close()
	}()

	// Event loop tracking inbound socket packets
	type WSMessage struct {
		ReceiverID string `json:"receiver_id"`
		Content    string `json:"content"`
	}

	for {
		var inbound WSMessage
		err := c.ReadJSON(&inbound)
		if err != nil {
			break // Client disconnected or payload error
		}

		// Route message through business logic layer
		msg, err := h.usecase.SendMessage(context.Background(), userID, inbound.ReceiverID, inbound.Content)
		if err != nil {
			_ = c.WriteJSON(fiber.Map{"error": err.Error()})
			continue
		}

		// Echo message back to sender to confirm acknowledgment
		_ = c.WriteJSON(msg)

		// Forward real-time stream execution if target peer is connected to WS instance
		h.hub.mu.RLock()
		if targetConn, online := h.hub.clients[inbound.ReceiverID]; online {
			_ = targetConn.WriteJSON(msg)
		}
		h.hub.mu.RUnlock()
	}
}