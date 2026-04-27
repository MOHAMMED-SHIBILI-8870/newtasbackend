package authhandler

import (
	"backend/internal/config"
	"backend/internal/entity"
	"backend/internal/usecase"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var input struct {
		Fullname string `json:"full_name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Phone    string `json:"phone"` // ✅ added
	}

	// Parse request
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Validation
	if input.Fullname == "" || input.Email == "" || len(input.Password) < 6 || input.Phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid input",
		})
	}

	var existUser entity.User

	// Check existing user
	if err := config.DB.Where("email = ?", input.Email).First(&existUser).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Email already registered",
		})
	}

	// Hash password
	hashpass, err := usecase.HashPassword(input.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error creating account",
		})
	}

	user := entity.User{
		FullName:     input.Fullname,
		Email:        input.Email,
		HashPassword: hashpass,
		Role:         "user",
		IsVerified:   false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save user
	if err := config.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not create user",
		})
	}

	// ✅ Send OTP via Twilio
	status, err := h.OTPUsecase.SendOTP(input.Phone)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to send OTP",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User registered successfully. OTP sent to your phone.",
		"status":  status,
	})
}
type AuthHandler struct {
	OTPUsecase *usecase.OTPUsecase
}

func NewAuthHandler(otp *usecase.OTPUsecase) *AuthHandler {
	return &AuthHandler{OTPUsecase: otp}
}

type SendOTPRequest struct {
	Phone string `json:"phone"`
}

type VerifyOTPRequest struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

// POST /send-otp
func (h *AuthHandler) SendOTP(c *fiber.Ctx) error {
	var req SendOTPRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	status, err := h.OTPUsecase.SendOTP(req.Phone)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": status,
	})
}

// POST /verify-otp
func (h *AuthHandler) VerifyOTP(c *fiber.Ctx) error {
	var req VerifyOTPRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	ok, err := h.OTPUsecase.VerifyOTP(req.Phone, req.Code)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if !ok {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid OTP",
		})
	}

	return c.JSON(fiber.Map{
		"message": "OTP verified",
	})
}