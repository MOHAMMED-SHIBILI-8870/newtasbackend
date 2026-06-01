package handler

import (
	"backend/internal/config"
	"backend/internal/dto"
	"backend/internal/entity"
	"backend/internal/response"
	"backend/internal/usecase"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func setAuthCookie(c *fiber.Ctx, name string, value string, exp time.Time) {
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    value,
		Expires:  exp,
		HTTPOnly: true,
		SameSite: "Lax",
		Secure:   false,
		Path:     "/",
	})
}

func userToResponse(user entity.User) dto.AuthUserResponse {
	return dto.AuthUserResponse{
		ID:         user.ID,
		FullName:   user.FullName,
		Email:      user.Email,
		Role:       usecase.NormalizeRole(user.Role),
		IsBlocked:  user.IsBlocked,
		IsVerified: user.IsVerified,
	}
}

func Register(c *fiber.Ctx) error {
	var input dto.RegisterRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	if input.FullName == "" || input.Email == "" || input.Password == "" {
		return response.Fail(c, fiber.StatusBadRequest, "all fields are required", nil)
	}
	if len(input.Password) < 6 {
		return response.Fail(c, fiber.StatusBadRequest, "password must be at least 6 characters", nil)
	}

	var existing entity.User
	if err := config.DB.Where("email = ?", input.Email).First(&existing).Error; err == nil {
		return response.Fail(c, fiber.StatusConflict, "email already registered", nil)
	}

	hashpass, err := usecase.HashPassword(input.Password)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to hash password", err)
	}

	user := entity.User{
		FullName:     input.FullName,
		Email:        input.Email,
		HashPassword: hashpass,
		Role:         usecase.NormalizeRole("user"),
		IsVerified:   false,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "could not create user", err)
	}

	_ = config.DB.Where("user_id = ? AND purpose = ?", user.ID, "signup").Delete(&entity.OTP{}).Error

	otp, err := usecase.CreateOTP(config.DB, input.Email, "signup", 5)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "could not generate OTP", err)
	}

	if err := usecase.SentOTPEmail(input.Email, otp, "signup"); err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "could not send OTP email", err)
	}

	return response.Success(c, fiber.StatusCreated, "user registered successfully, OTP sent to email", fiber.Map{
		"user": userToResponse(user),
	})
}

func VerifyOTPHandler(c *fiber.Ctx) error {
	var input dto.OTPRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	if input.Email == "" || input.OTP == "" || input.Purpose == "" {
		return response.Fail(c, fiber.StatusBadRequest, "missing required fields", nil)
	}

	var user entity.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return response.Fail(c, fiber.StatusNotFound, "user not found", err)
	}

	valid, err := usecase.VerifyOTP(input.Email, input.OTP, input.Purpose)
	if err != nil || !valid {
		return response.Fail(c, fiber.StatusBadRequest, "invalid or expired OTP", err)
	}

	if input.Purpose == "signup" {
		_ = config.DB.Model(&user).Updates(map[string]any{
			"is_verified": true,
			"updated_at":  time.Now(),
		}).Error
	}

	_ = config.DB.Where("user_id = ? AND purpose = ?", user.ID, input.Purpose).Delete(&entity.OTP{}).Error

	return response.Success(c, fiber.StatusOK, "OTP verified successfully", nil)
}

func Login(c *fiber.Ctx) error {
	var input dto.LoginRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	if input.Email == "" || input.Password == "" {
		return response.Fail(c, fiber.StatusBadRequest, "email and password required", nil)
	}

	var user entity.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return response.Fail(c, fiber.StatusNotFound, "email not found", err)
	}

	if !user.IsVerified {
		return response.Fail(c, fiber.StatusUnauthorized, "user not verified", nil)
	}
	if user.IsBlocked {
		return response.Fail(c, fiber.StatusForbidden, "user blocked", nil)
	}
	if !usecase.Checkpassword(input.Password, user.HashPassword) {
		return response.Fail(c, fiber.StatusUnauthorized, "wrong password", nil)
	}

	accessToken, err := usecase.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to generate access token", err)
	}

	refreshToken, hashedToken, err := usecase.GenerateRefreshToken()
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to generate refresh token", err)
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := usecase.SaveRefreshToken(config.DB, user.ID, hashedToken, expiresAt); err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to save refresh token", err)
	}

	setAuthCookie(c, "access_token", accessToken, time.Now().Add(15*time.Minute))
	setAuthCookie(c, "refresh_token", refreshToken, expiresAt)

	return response.Success(c, fiber.StatusOK, "login successful", dto.AuthResponse{
		AccessToken: accessToken,
		User:        userToResponse(user),
	})
}

func ForgetPassword(c *fiber.Ctx) error {
	var input dto.ForgotPasswordRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	if input.Email == "" {
		return response.Fail(c, fiber.StatusBadRequest, "email is required", nil)
	}

	var user entity.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return response.Fail(c, fiber.StatusNotFound, "user not found", err)
	}

	_ = config.DB.Where("user_id = ? AND purpose = ?", user.ID, "reset_password").Delete(&entity.OTP{}).Error

	otp, err := usecase.CreateOTP(config.DB, input.Email, "reset_password", 5)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "could not generate OTP", err)
	}

	if err := usecase.SentOTPEmail(input.Email, otp, "reset_password"); err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "could not send OTP email", err)
	}

	return response.Success(c, fiber.StatusOK, "OTP sent successfully", nil)
}

func ResetPassword(c *fiber.Ctx) error {
	var input dto.ResetPasswordRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	if input.Email == "" || input.NewPassword == "" || input.OTP == "" {
		return response.Fail(c, fiber.StatusBadRequest, "missing required fields", nil)
	}
	if len(input.NewPassword) < 6 {
		return response.Fail(c, fiber.StatusBadRequest, "password must be at least 6 characters", nil)
	}

	var user entity.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return response.Fail(c, fiber.StatusNotFound, "user not found", err)
	}

	valid, err := usecase.VerifyOTP(input.Email, input.OTP, "reset_password")
	if err != nil || !valid {
		return response.Fail(c, fiber.StatusBadRequest, "invalid or expired OTP", err)
	}

	hashedPass, err := usecase.HashPassword(input.NewPassword)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to hash password", err)
	}

	if err := config.DB.Model(&entity.User{}).
		Where("email = ?", input.Email).
		Updates(map[string]any{
			"hash_password": hashedPass,
			"updated_at":    time.Now(),
		}).Error; err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to update password", err)
	}

	_ = config.DB.Where("user_id = ? AND purpose = ?", user.ID, "reset_password").Delete(&entity.OTP{}).Error

	return response.Success(c, fiber.StatusOK, "password reset successfully", nil)
}

func ResendOtpHandler(c *fiber.Ctx) error {
	var input dto.ResendOTPRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	if input.Email == "" || input.Purpose == "" {
		return response.Fail(c, fiber.StatusBadRequest, "missing required fields", nil)
	}

	var user entity.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return response.Fail(c, fiber.StatusNotFound, "user not found", err)
	}

	if user.IsVerified && input.Purpose == "signup" {
		return response.Fail(c, fiber.StatusBadRequest, "user already verified", nil)
	}

	_ = config.DB.Where("user_id = ? AND purpose = ?", user.ID, input.Purpose).Delete(&entity.OTP{}).Error

	otp, err := usecase.CreateOTP(config.DB, input.Email, input.Purpose, 5)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "could not generate OTP", err)
	}

	if err := usecase.SentOTPEmail(input.Email, otp, input.Purpose); err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "could not send OTP email", err)
	}

	return response.Success(c, fiber.StatusOK, "OTP resent successfully", nil)
}

func Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken != "" {
		_ = usecase.DeleteReToken(config.DB, refreshToken)
	}

	setAuthCookie(c, "access_token", "", time.Now().Add(-time.Hour))
	setAuthCookie(c, "refresh_token", "", time.Now().Add(-time.Hour))

	return response.Success(c, fiber.StatusOK, "logged out successfully", nil)
}

func RefreshTokenHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		refreshToken := c.Cookies("refresh_token")
		if refreshToken == "" {
			return response.Fail(c, fiber.StatusUnauthorized, "missing refresh token", nil)
		}

		rt, err := usecase.ValidateRefreshToken(db, refreshToken)
		if err != nil {
			return response.Fail(c, fiber.StatusUnauthorized, "invalid or expired refresh token", err)
		}

		_ = usecase.DeleteReToken(db, refreshToken)

		var user entity.User
		if err := db.First(&user, rt.UserID).Error; err != nil {
			return response.Fail(c, fiber.StatusNotFound, "user not found", err)
		}

		newAccessToken, err := usecase.GenerateAccessToken(user.ID, user.Role)
		if err != nil {
			return response.Fail(c, fiber.StatusInternalServerError, "failed to generate access token", err)
		}

		newRefreshToken, hashedToken, err := usecase.GenerateRefreshToken()
		if err != nil {
			return response.Fail(c, fiber.StatusInternalServerError, "failed to generate refresh token", err)
		}

		expiresAt := time.Now().Add(7 * 24 * time.Hour)
		if err := usecase.SaveRefreshToken(db, user.ID, hashedToken, expiresAt); err != nil {
			return response.Fail(c, fiber.StatusInternalServerError, "failed to save refresh token", err)
		}

		setAuthCookie(c, "access_token", newAccessToken, time.Now().Add(15*time.Minute))
		setAuthCookie(c, "refresh_token", newRefreshToken, expiresAt)

		return response.Success(c, fiber.StatusOK, "token refreshed successfully", dto.AuthResponse{
			AccessToken: newAccessToken,
			User:        userToResponse(user),
		})
	}
}
