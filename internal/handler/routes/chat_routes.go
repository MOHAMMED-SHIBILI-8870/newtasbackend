package routes

import (
	"backend/internal/handler"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func MapChatRoutes(app *fiber.App, chatHandler *handler.ChatHandler, authMiddleware fiber.Handler) {
	api := app.Group("/api/v1/chat", authMiddleware) 

	api.Post("/message", chatHandler.RESTSendMessage)
	api.Get("/history", chatHandler.RESTFetchHistory)
	api.Get("/ws", websocket.New(chatHandler.WebSocketHandler))
}