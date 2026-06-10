package routes

import (
	"backend/internal/handler"
	"backend/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func GuideRoutes(app fiber.Router,h *handler.GuideHandler,auth fiber.Handler) {

	guide := app.Group("/guide")

	guide.Use(auth)
	guide.Use(middleware.RoleMiddleware("guide"))

	guide.Get("/profile", h.GetProfile)
	guide.Put("/profile", h.UpdateProfile)
}