package routes

import (
	"backend/internal/handler"
	"github.com/gofiber/fiber/v2"
)

func VerificationRoutes(app fiber.Router, verificationHandler *handler.VerificationHandler, authMiddleware fiber.Handler) {
	app.Post("/verifications", authMiddleware, verificationHandler.SubmitVerification)
	app.Get("/verifications/:bookingId",authMiddleware,verificationHandler.GetVerification)
}
