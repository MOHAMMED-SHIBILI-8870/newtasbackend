package routes

import (
	"backend/internal/handler"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AuthRoutes(app *fiber.App, db *gorm.DB) {
	auth := app.Group("/auth")

	auth.Post("/register", handler.Register)
	auth.Post("/login", handler.Login)
	auth.Post("/verify-otp", handler.VerifyOTPHandler)
	auth.Post("/forgot-password", handler.ForgetPassword)
	auth.Post("/reset-password", handler.ResetPassword)
	auth.Post("/resend-otp", handler.ResendOtpHandler)
	auth.Post("/logout", handler.Logout)
	auth.Post("/refresh", handler.RefreshTokenHandler(db))
}
