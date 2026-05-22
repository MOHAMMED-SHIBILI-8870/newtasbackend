// internal/handler/routes/ai_routes.go
package routes

import (
	"backend/internal/handler"

	"github.com/gofiber/fiber/v2"
)

func SetupAIRoutes(router fiber.Router, aiHandler *handler.AIHandler, auth fiber.Handler, admin fiber.Handler) {
	aiGroup := router.Group("/ai", auth)

	aiGroup.Post("/chat", aiHandler.GenerateTripPlan)
	aiGroup.Post("/requests", aiHandler.CreateTripRequest)
	aiGroup.Get("/requests", aiHandler.GetMyTripRequests)

	adminAIGroup := router.Group("/admin/ai-requests", auth, admin)
	adminAIGroup.Get("/", aiHandler.GetAllTripRequests)
	adminAIGroup.Patch("/:id/approve", aiHandler.ApproveTripRequest)
	adminAIGroup.Patch("/:id/reject", aiHandler.RejectTripRequest)
}
