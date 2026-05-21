package handler

import (
	"backend/internal/config"
	"backend/internal/entity"
	"backend/internal/usecase"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ========================================
// Helper Function
// ========================================

func setAuthCookie(
	c *fiber.Ctx,
	name string,
	value string,
	exp time.Time,
) {
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    value,
		Expires:  exp,
		HTTPOnly: true,
		SameSite: "None",
		Secure:   false, // true in production with HTTPS
		Path:     "/",
	})
}

// ========================================
// Register
// ========================================

func Register(c *fiber.Ctx) error {
	var input struct {
		Fullname string `json:"full_name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// validation
	if input.Fullname == "" || input.Email == "" || input.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "all fields are required",
		})
	}

	if len(input.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "password must be at least 6 characters",
		})
	}

	var existUser entity.User
	if err := config.DB.Where("email = ?", input.Email).First(&existUser).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "email already registered",
		})
	}

	hashpass, err := usecase.HashPassword(input.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to hash password",
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

	// delete old OTP if exists
	config.DB.Where("user_id = ? AND purpose = ?", user.ID, "signup").
		Delete(&entity.OTP{})

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
		"message": "user registered successfully, OTP sent to email",
	})
}

// ========================================
// Verify OTP
// ========================================

func VerifyOTPHandler(c *fiber.Ctx) error {
	var input struct {
		Email   string `json:"email"`
		OTP     string `json:"otp"`
		Purpose string `json:"purpose"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if input.Email == "" || input.OTP == "" || input.Purpose == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing required fields",
		})
	}

	var user entity.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	valid, err := usecase.VerifyOTP(input.Email, input.OTP, input.Purpose)
	if err != nil || !valid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid or expired OTP",
		})
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

// ========================================
// Login
// ========================================

func Login(c *fiber.Ctx) error {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if input.Email == "" || input.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email and password required",
		})
	}

	var user entity.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "email not found",
		})
	}

	if !user.IsVerified {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user not verified",
		})
	}

	if user.IsBlocked {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "user blocked",
		})
	}

	if !usecase.Checkpassword(input.Password, user.HashPassword) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "wrong password",
		})
	}

	accessToken, err := usecase.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate access token",
		})
	}

	refreshToken, hashedToken, err := usecase.GenerateRefreshToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate refresh token",
		})
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	if err := usecase.SaveRefreshToken(
		config.DB,
		user.ID,
		hashedToken,
		expiresAt,
	); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to save refresh token",
		})
	}

	// set cookies
	setAuthCookie(
		c,
		"access_token",
		accessToken,
		time.Now().Add(15*time.Minute),
	)

	setAuthCookie(
		c,
		"refresh_token",
		refreshToken,
		expiresAt,
	)

	return c.JSON(fiber.Map{
	"status":       "logged in",
	"access_token": accessToken,
	"user": fiber.Map{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	},
})
}

// ========================================
// Forget Password
// ========================================

func ForgetPassword(c *fiber.Ctx) error {
	var input struct {
		Email string `json:"email"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if input.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "email is required",
		})
	}

	var user entity.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	// delete old OTP
	config.DB.Where("user_id = ? AND purpose = ?", user.ID, "reset_password").
		Delete(&entity.OTP{})

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
		"msg":    "OTP sent successfully",
	})
}

// ========================================
// Reset Password
// ========================================

func ResetPassword(c *fiber.Ctx) error {
	var input struct {
		Email       string `json:"email"`
		NewPassword string `json:"new_password"`
		OTP         string `json:"otp"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if input.Email == "" || input.NewPassword == "" || input.OTP == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing required fields",
		})
	}

	if len(input.NewPassword) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "password must be at least 6 characters",
		})
	}

	var user entity.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	valid, err := usecase.VerifyOTP(
		input.Email,
		input.OTP,
		"reset_password",
	)

	if err != nil || !valid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid or expired OTP",
		})
	}

	hashedPass, err := usecase.HashPassword(input.NewPassword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to hash password",
		})
	}

	if err := config.DB.Model(&entity.User{}).
		Where("email = ?", input.Email).
		Updates(map[string]interface{}{
			"hash_password": hashedPass,
			"updated_at":    time.Now(),
		}).Error; err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update password",
		})
	}

	config.DB.Where("user_id = ? AND purpose = ?", user.ID, "reset_password").
		Delete(&entity.OTP{})

	return c.JSON(fiber.Map{
		"status": "success",
		"msg":    "password reset successfully",
	})
}

// ========================================
// Resend OTP
// ========================================

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
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	if user.IsVerified && input.Purpose == "signup" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user already verified",
		})
	}

	// delete old OTP
	config.DB.Where("user_id = ? AND purpose = ?", user.ID, input.Purpose).
		Delete(&entity.OTP{})

	otp, err := usecase.CreateOTP(
		config.DB,
		input.Email,
		input.Purpose,
		5,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not generate OTP",
		})
	}

	if err := usecase.SentOTPEmail(
		input.Email,
		otp,
		input.Purpose,
	); err != nil {

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not send OTP email",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "OTP resent successfully",
	})
}

// ========================================
// Logout
// ========================================

func Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")

	if refreshToken != "" {
		_ = usecase.DeleteReToken(config.DB, refreshToken)
	}

	// clear access token cookie
	setAuthCookie(
		c,
		"access_token",
		"",
		time.Now().Add(-time.Hour),
	)

	// clear refresh token cookie
	setAuthCookie(
		c,
		"refresh_token",
		"",
		time.Now().Add(-time.Hour),
	)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "logged out successfully",
	})
}

// ========================================
// Refresh Token
// ========================================

func RefreshTokenHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		refreshToken := c.Cookies("refresh_token")

		if refreshToken == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing refresh token",
			})
		}

		rt, err := usecase.ValidateRefreshToken(db, refreshToken)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired refresh token",
			})
		}

		// rotate refresh token
		_ = usecase.DeleteReToken(db, refreshToken)

		var user entity.User
		if err := db.First(&user, rt.UserId).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "user not found",
			})
		}

		newAccessToken, err := usecase.GenerateAccessToken(
			user.ID,
			user.Role,
		)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to generate access token",
			})
		}

		newRefreshToken, hashedToken, err := usecase.GenerateRefreshToken()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to generate refresh token",
			})
		}

		expiresAt := time.Now().Add(7 * 24 * time.Hour)

		if err := usecase.SaveRefreshToken(
			db,
			user.ID,
			hashedToken,
			expiresAt,
		); err != nil {

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to save refresh token",
			})
		}

		// set new cookies
		setAuthCookie(
			c,
			"access_token",
			newAccessToken,
			time.Now().Add(15*time.Minute),
		)

		setAuthCookie(
			c,
			"refresh_token",
			newRefreshToken,
			expiresAt,
		)

		return c.JSON(fiber.Map{
			"message": "token refreshed successfully",
		})
	}
}