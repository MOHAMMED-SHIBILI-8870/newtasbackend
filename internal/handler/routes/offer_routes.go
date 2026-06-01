package routes

import (
	"backend/internal/handler"

	"github.com/gofiber/fiber/v2"
)

func OfferRoutes(app fiber.Router, h *handler.OfferHandler, auth fiber.Handler, permission func(...string) fiber.Handler) {
	offers := app.Group("/offers", auth)
	offers.Get("/active", h.ListActiveOffers)
	offers.Post("/validate", h.ValidateCoupon)

	adminOffers := app.Group("/admin/offers", auth)
	adminOffers.Get("/", permission("manage_offers"), h.ListOffers)
	adminOffers.Post("/", permission("manage_offers"), h.CreateOffer)
	adminOffers.Put("/:id", permission("manage_offers"), h.UpdateOffer)
	adminOffers.Delete("/:id", permission("manage_offers"), h.DeleteOffer)
}
