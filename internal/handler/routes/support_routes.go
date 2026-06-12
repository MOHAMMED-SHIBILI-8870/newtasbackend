package routes

import (
	"backend/internal/handler"

	"github.com/gofiber/fiber/v2"
)

func SupportRoutes(app fiber.Router, h *handler.SupportHandler, auth fiber.Handler) {
	userGroup := app.Group("/support", auth)
	userGroup.Post("/requests", h.CreateRequest)
	userGroup.Get("/requests", h.ListMyRequests)

	adminGroup := app.Group("/admin/support", auth)
	adminGroup.Get("/requests", h.ListAllRequests)
	adminGroup.Patch("/requests/:id/assign", h.AssignAgent)
	adminGroup.Patch("/requests/:id/status", h.UpdateStatus)
}
