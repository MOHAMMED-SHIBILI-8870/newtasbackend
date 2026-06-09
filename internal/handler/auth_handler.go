package handler

import (
	"backend/internal/dto"
	"backend/internal/entity"
	"backend/internal/response"
	"backend/internal/usecase"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

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

func authTokenStatusFromErr(err error) int {
	if err == nil {
		return fiber.StatusUnauthorized
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "blocked"):
		return fiber.StatusForbidden
	case strings.Contains(msg, "not verified"):
		return fiber.StatusForbidden
	case strings.Contains(msg, "not found"):
		return fiber.StatusNotFound
	default:
		return fiber.StatusUnauthorized
	}
}

func refreshTokenStatusFromErr(err error) int {
	if err == nil {
		return fiber.StatusUnauthorized
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "blocked"):
		return fiber.StatusForbidden
	case strings.Contains(msg, "not verified"):
		return fiber.StatusForbidden
	default:
		return fiber.StatusUnauthorized
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
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
	if err := h.db.Where("email = ?", input.Email).First(&existing).Error; err == nil {
		return response.Fail(c, fiber.StatusConflict, "email already registered", nil)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to check existing user", err)
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

	if err := h.db.Create(&user).Error; err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "could not create user", err)
	}

	otp, err := usecase.CreateOTP(h.db, c.Context(), input.Email, "signup", 5)
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

func (h *AuthHandler) VerifyOTPHandler(c *fiber.Ctx) error {
	var input dto.OTPRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	if input.Email == "" || input.OTP == "" || input.Purpose == "" {
		return response.Fail(c, fiber.StatusBadRequest, "missing required fields", nil)
	}

	var user entity.User
	if err := h.db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return response.Fail(c, fiber.StatusNotFound, "user not found", err)
	}

	valid, err := usecase.VerifyOTP(h.db, c.Context(), input.Email, input.OTP, input.Purpose)
	if err != nil || !valid {
		return response.Fail(c, fiber.StatusBadRequest, "invalid or expired OTP", err)
	}

	return response.Success(c, fiber.StatusOK, "OTP verified successfully", nil)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var input dto.LoginRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	if input.Email == "" || input.Password == "" {
		return response.Fail(c, fiber.StatusBadRequest, "email and password required", nil)
	}

	var user entity.User
	if err := h.db.Where("email = ?", input.Email).First(&user).Error; err != nil {
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
	if err := usecase.SaveRefreshToken(h.db, c.Context(), user.ID, hashedToken, expiresAt); err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to save refresh token", err)
	}

	setAuthCookie(c, "access_token", accessToken, time.Now().Add(15*time.Minute))
	setAuthCookie(c, "refresh_token", refreshToken, expiresAt)

	return response.Success(c, fiber.StatusOK, "login successful", dto.AuthResponse{
		AccessToken: accessToken,
		User:        userToResponse(user),
	})
}

func (h *AuthHandler) ForgetPassword(c *fiber.Ctx) error {
	var input dto.ForgotPasswordRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	if input.Email == "" {
		return response.Fail(c, fiber.StatusBadRequest, "email is required", nil)
	}

	var user entity.User
	if err := h.db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return response.Fail(c, fiber.StatusNotFound, "user not found", err)
	}

	otp, err := usecase.CreateOTP(h.db, c.Context(), input.Email, "reset_password", 5)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "could not generate OTP", err)
	}

	if err := usecase.SentOTPEmail(input.Email, otp, "reset_password"); err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "could not send OTP email", err)
	}

	return response.Success(c, fiber.StatusOK, "OTP sent successfully", nil)
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
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
	if err := h.db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return response.Fail(c, fiber.StatusNotFound, "user not found", err)
	}

	valid, err := usecase.VerifyOTP(h.db, c.Context(), input.Email, input.OTP, "reset_password")
	if err != nil || !valid {
		return response.Fail(c, fiber.StatusBadRequest, "invalid or expired OTP", err)
	}

	hashedPass, err := usecase.HashPassword(input.NewPassword)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to hash password", err)
	}

	if err := h.db.Model(&entity.User{}).
		Where("email = ?", input.Email).
		Updates(map[string]any{
			"hash_password": hashedPass,
			"updated_at":    time.Now(),
		}).Error; err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to update password", err)
	}

	// Password reset invalidates every existing refresh token so stolen sessions cannot persist.
	if err := usecase.RevokeRefreshTokensByUserID(h.db, c.Context(), user.ID); err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to revoke refresh tokens", err)
	}

	return response.Success(c, fiber.StatusOK, "password reset successfully", nil)
}

func (h *AuthHandler) ResendOtpHandler(c *fiber.Ctx) error {
	var input dto.ResendOTPRequest
	if err := c.BodyParser(&input); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "invalid request body", err)
	}

	if input.Email == "" || input.Purpose == "" {
		return response.Fail(c, fiber.StatusBadRequest, "missing required fields", nil)
	}

	var user entity.User
	if err := h.db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return response.Fail(c, fiber.StatusNotFound, "user not found", err)
	}

	if user.IsVerified && input.Purpose == "signup" {
		return response.Fail(c, fiber.StatusBadRequest, "user already verified", nil)
	}

	otp, err := usecase.CreateOTP(h.db, c.Context(), input.Email, input.Purpose, 5)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "could not generate OTP", err)
	}

	if err := usecase.SentOTPEmail(input.Email, otp, input.Purpose); err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "could not send OTP email", err)
	}

	return response.Success(c, fiber.StatusOK, "OTP resent successfully", nil)
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken != "" {
		_ = usecase.DeleteReToken(h.db, c.Context(), refreshToken)
	}

	setAuthCookie(c, "access_token", "", time.Now().Add(-time.Hour))
	setAuthCookie(c, "refresh_token", "", time.Now().Add(-time.Hour))

	return response.Success(c, fiber.StatusOK, "logged out successfully", nil)
}

func (h *AuthHandler) RefreshTokenHandler(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return response.Fail(c, fiber.StatusUnauthorized, "missing refresh token", nil)
	}

	user, newRefreshToken, expiresAt, err := usecase.RotateRefreshToken(h.db, c.Context(), refreshToken)
	if err != nil {
		return response.Fail(c, refreshTokenStatusFromErr(err), "invalid or expired refresh token", err)
	}

	newAccessToken, err := usecase.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "failed to generate access token", err)
	}

	setAuthCookie(c, "access_token", newAccessToken, time.Now().Add(15*time.Minute))
	setAuthCookie(c, "refresh_token", newRefreshToken, expiresAt)

	return response.Success(c, fiber.StatusOK, "token refreshed successfully", dto.AuthResponse{
		AccessToken: newAccessToken,
		User:        userToResponse(*user),
	})
}
