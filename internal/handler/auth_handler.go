package handler

import (
	"backend/internal/config"
	"backend/internal/entity"
	"backend/internal/usecase"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Register
func Register(c *fiber.Ctx) error {
	var input struct {
		Fullname string `json:"full_name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var existUser entity.User
	if err := config.DB.Where("email = ?", input.Email).First(&existUser).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Email already registered",
		})
	}

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

	if err := config.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not create user",
		})
	}

	otp, err := usecase.CreateOTP(config.DB, input.Email, "signup", 5)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not generate OTP",
		})
	}

	if err := usecase.SentOTPEmail(input.Email, otp, "signup"); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not send OTP email",
		})
	}

	return c.JSON(fiber.Map{
		"message": "User registered successfully. OTP sent to your email.",
	})
}

// Verify OTP
func VerifyOTPHandler(c *fiber.Ctx) error {
	var input struct {
		Email   string `json:"email"`
		OTP     string `json:"otp"`
		Purpose string `json:"purpose"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var user entity.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	valid, err := usecase.VerifyOTP(input.Email, input.OTP, input.Purpose)
	if err != nil || !valid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "OTP is expired or wrong"})
	}

	if input.Purpose == "signup" {
		config.DB.Model(&user).Updates(map[string]interface{}{
			"is_verified": true,
			"updated_at":  time.Now(),
		})
	}

	config.DB.Where("user_id = ? AND purpose = ?", user.ID, input.Purpose).
		Delete(&entity.OTP{})

	return c.JSON(fiber.Map{
		"message": "OTP verified successfully",
	})
}

// Login
func Login(c *fiber.Ctx) error {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var users entity.User
	if err := config.DB.Where("email = ?", input.Email).First(&users).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "email not found"})
	}

	if !users.IsVerified {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "user not verified"})
	}

	if users.IsBlocked {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "user blocked"})
	}

	if !usecase.Checkpassword(input.Password, users.HashPassword) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "wrong password"})
	}

	accessToken, _ := usecase.GenerateAccessToken(users.ID, users.Role)
	refreshToken, hashedToken, _ := usecase.GenerateRefreshToken()

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	usecase.SaveRefreshToken(config.DB, users.ID, hashedToken, expiresAt)

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HTTPOnly: true,
		SameSite: "None", // 🔥 REQUIRED for cross-origin
		Secure:   false,  // true in production (HTTPS)
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  expiresAt,
		HTTPOnly: true,
	})

	return c.JSON(fiber.Map{
		"status":       "logged in",
		"role":         users.Role,
		"access_token": accessToken,
	})
}

func ForgetPassword(c *fiber.Ctx) error {
	var input struct {
		Email string `json:"email"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	// manual validation (Fiber doesn't support `binding`)
	if input.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email is required",
		})
	}

	var user entity.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("user not found")
	}

	otp, err := usecase.CreateOTP(config.DB, input.Email, "reset_password", 5)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not generate OTP",
		})
	}

	if err := usecase.SentOTPEmail(input.Email, otp, "reset_password"); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not send OTP email",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"msg":    "OTP sent to your email for password reset",
	})
}

func ResetPassword(c *fiber.Ctx) error {
	var input struct {
		Email       string `json:"email"`
		NewPassword string `json:"new_password"`
		OTP         string `json:"otp"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	if input.Email == "" || input.NewPassword == "" || input.OTP == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing required fields",
		})
	}

	var user entity.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	valid, err := usecase.VerifyOTP(input.Email, input.OTP, "reset_password")
	if !valid || err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid or expired token")
	}

	hashedPass, err := usecase.HashPassword(input.NewPassword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	if err := config.DB.Model(entity.User{}).
		Where("email = ?", user.Email).
		Updates(map[string]interface{}{
			"hash_password": hashedPass,
			"updated_at":    time.Now(),
		}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	config.DB.Where("user_id = ? AND purpose = ?", user.ID, "reset_password").
		Delete(&entity.OTP{})

	return c.JSON(fiber.Map{
		"status": "success",
		"msg":    "Password reset successfully",
	})
}

func ResendOtpHandler(c *fiber.Ctx) error {
	var input struct {
		Email   string `json:"email"`
		Purpose string `json:"purpose"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if input.Email == "" || input.Purpose == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing required fields",
		})
	}

	var user entity.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid credentials",
		})
	}

	if user.IsVerified && input.Purpose == "signup" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user already verified",
		})
	}

	otp, err := usecase.CreateOTP(config.DB, input.Email, input.Purpose, 5)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not generate OTP",
		})
	}

	if err := usecase.SentOTPEmail(input.Email, otp, input.Purpose); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not send OTP email",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "OTP resent successfully. Please check your email",
	})
}

func Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")

	if refreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Refresh token required",
		})
	}

	if err := usecase.DeleteReToken(config.DB, refreshToken); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	// clear cookie
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
	})

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Logged out successfully",
	})
}

// create new access token and create new refresh token
func RefreshTokenHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// ✅ get refresh token from cookie
		refreshToken := c.Cookies("refresh_token")
		if refreshToken == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "missing refresh token",
			})
		}

		// 🔍 validate refresh token
		rt, err := usecase.ValidateRefreshToken(db, refreshToken)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "invalid or expired refresh token",
			})
		}

		// ❌ delete old refresh token (rotation)
		_ = usecase.DeleteReToken(db, refreshToken)

		// 👤 get user from DB
		var user entity.User
		if err := db.First(&user, rt.UserId).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "user not found",
			})
		}

		// 🔐 generate new access token
		newAccessToken, err := usecase.GenerateAccessToken(user.ID, user.Role)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "failed to generate access token",
			})
		}

		// 🔄 generate new refresh token
		newRefreshToken, hashed, err := usecase.GenerateRefreshToken()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "failed to generate refresh token",
			})
		}

		// 💾 save new refresh token
		expiresAt := time.Now().Add(7 * 24 * time.Hour)
		if err := usecase.SaveRefreshToken(db, user.ID, hashed, expiresAt); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "failed to save refresh token",
			})
		}

		// 🍪 overwrite access token cookie
		c.Cookie(&fiber.Cookie{
			Name:     "access_token",
			Value:    newAccessToken,
			HTTPOnly: true,
			SameSite: "None",
			Secure:   false,
			Expires:  time.Now().Add(15 * time.Minute),
		})

		// 🍪 overwrite refresh token cookie
		c.Cookie(&fiber.Cookie{
			Name:     "refresh_token",
			Value:    newRefreshToken,
			HTTPOnly: true,
			SameSite: "None",
			Secure:   false,
			Expires:  expiresAt,
		})

		return c.JSON(fiber.Map{
			"message": "token refreshed successfully",
			"access_token":newAccessToken,
			"refresh_token":newRefreshToken,
		})
	}
}