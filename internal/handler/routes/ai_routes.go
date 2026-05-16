// internal/handler/routes/ai_routes.go
package routes

import (
	"backend/internal/handler"
	"github.com/gofiber/fiber/v2"
)

func SetupAIRoutes(router fiber.Router, aiHandler *handler.AIHandler) {
	aiGroup := router.Group("/ai")

	aiGroup.Post("/chat", aiHandler.CustomTripChat)
}