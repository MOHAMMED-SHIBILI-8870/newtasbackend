// internal/handler/routes/ai_routes.go
package routes

import (
	"backend/internal/handler"
	"backend/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupAIRoutes(router fiber.Router, aiHandler *handler.AIHandler) {
	aiGroup := router.Group("/ai",middleware.AuthMiddleware())

	aiGroup.Post("/chat", aiHandler.GenerateTripPlan)
}