package handler

import (
	"backend/internal/entity"
	"context"
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

	messages, err := h.usecase.GetMessages(context.Background(), userID, guideID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"messages": messages})
}

// WebSocket connection handler
func (h *ChatHandler) WebSocketHandler(c *websocket.Conn) {
	userID := c.Query("user_id")
	if userID == "" {
		_ = c.WriteJSON(fiber.Map{"error": "unauthorized connection target missing query param: user_id"})
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