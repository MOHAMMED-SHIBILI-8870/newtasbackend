package routes

import (
	"backend/internal/handler"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AuthRoutes(app *fiber.App, db *gorm.DB) {
	auth := app.Group("/auth")
	authHandler := handler.NewAuthHandler(db)

	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/verify-otp", authHandler.VerifyOTPHandler)
	auth.Post("/forgot-password", authHandler.ForgetPassword)
	auth.Post("/reset-password", authHandler.ResetPassword)
	auth.Post("/resend-otp", authHandler.ResendOtpHandler)
	auth.Post("/logout", authHandler.Logout)
	auth.Post("/refresh", authHandler.RefreshTokenHandler)
}
