package routes

import (
	"backend/internal/handler"

	"github.com/gofiber/fiber/v2"
)

func PaymentRoutes(app fiber.Router, h *handler.PaymentHandler,auth fiber.Handler){
	payment := app.Group("/payments",auth)

	payment.Post("/booking/:booking_id/advance",h.CreateAdvancePayment)
	payment.Post("/booking/:booking_id/advance/verify",h.VerifyAdvancePayment)
	payment.Post("/booking/:booking_id/balance",h.CreateBalancePayment)
	payment.Post("/booking/:booking_id/balance/verify",h.VerifyBalancePayment)
}