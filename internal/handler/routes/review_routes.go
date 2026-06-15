package routes

import (
	"backend/internal/handler"

	"github.com/gofiber/fiber/v2"
)

func ReviewRoutes(app fiber.Router, h *handler.ReviewHandler, auth fiber.Handler, permission func(...string) fiber.Handler) {
	reviews := app.Group("/reviews", auth)
	reviews.Post("/", h.CreateReview)
	reviews.Get("/me", h.ListMyReviews)
	reviews.Get("/trip/:trip_id", h.ListTripReviews)
	reviews.Get("/trip/:trip_id/summary", h.GetTripSummary)
	reviews.Get("/assigned", h.ListAssignedReviews)

	adminReviews := app.Group("/admin/reviews", auth)
	adminReviews.Get("/", permission("manage_reviews"), h.ListAllReviews)
	adminReviews.Delete("/:id", permission("manage_reviews"), h.DeleteReview)
}
