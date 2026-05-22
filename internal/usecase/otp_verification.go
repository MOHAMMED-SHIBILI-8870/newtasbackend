package usecase

import (
	"backend/internal/config"
	"backend/internal/entity"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ===============================
// Generate OTP
// ===============================
func GenerateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	otp := 100000 + n.Int64()
	return fmt.Sprintf("%06d", otp), nil
}

// ===============================
// Hash OTP
// ===============================
func HashOTP(otp string) string {
	hash := sha256.Sum256([]byte(otp))
	return hex.EncodeToString(hash[:])
}

// ===============================
// CREATE OTP
// ===============================
func CreateOTP(db *gorm.DB, email string, purpose string, expiryMinutes int) (string, error) {

	otp, err := GenerateOTP()
	if err != nil {
		return "", err
	}

	otpHash := HashOTP(otp)

	tx := db.Begin()

	// invalidate old OTPs
	if err := tx.Model(&entity.OTP{}).
		Where("email = ? AND purpose = ? AND is_used = false", email, purpose).
		Update("is_used", true).Error; err != nil {
		tx.Rollback()
		return "", err
	}

	newOTP := entity.OTP{
		Email:     email,
		OTPCode:   otpHash,
		Purpose:   purpose,
		IsUsed:    false,
		ExpiresAt: time.Now().Add(time.Minute * time.Duration(expiryMinutes)),
	}

	if err := tx.Create(&newOTP).Error; err != nil {
		tx.Rollback()
		return "", err
	}

	tx.Commit()

	// return RAW OTP (for sending via email)
	return otp, nil
}

// ===============================
// VERIFY OTP
// ===============================
func VerifyOTP(email string, otp string, purpose string) (bool, error) {

	err := config.DB.Transaction(func(tx *gorm.DB) error {

		var record entity.OTP
		hashed := HashOTP(otp)

		// find OTP
		if err := tx.Where(
			"email = ? AND purpose = ? AND otp_code = ? AND is_used = false AND expires_at > ?",
			email, purpose, hashed, time.Now(),
		).
			Order("created_at DESC").
			First(&record).Error; err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("invalid or expired OTP")
			}
			return err
		}

		// mark OTP used
		if err := tx.Model(&record).
			Update("is_used", true).Error; err != nil {
			return err
		}

		// OPTIONAL: verify user (if signup)
		if purpose == "signup" {
			if err := tx.Model(&entity.User{}).
				Where("email = ?", email).
				Update("is_verified", true).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return false, err
	}

	return true, nil
}

func VerifyOTPHandler(c *fiber.Ctx) error {
	var body struct {
		Email   string `json:"email"`
		OTP     string `json:"otp"`
		Purpose string `json:"purpose"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	ok, err := VerifyOTP(body.Email, body.OTP, body.Purpose)
	if err != nil {
		return c.Status(400).JSON(err.Error())
	}

	return c.JSON(fiber.Map{
		"success": ok,
		"message": "OTP verified successfully",
	})
}
