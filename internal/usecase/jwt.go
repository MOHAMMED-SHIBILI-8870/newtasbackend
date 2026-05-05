package usecase

import (
	"backend/internal/entity"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

//  Generate Access Token
func GenerateAccessToken(userID uint, role string) (string, error) {
	secretKey := os.Getenv("JWT_SECRETKEY")
	if secretKey == "" {
		return "", errors.New("JWT secret key not set")
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// 🔄 Generate Refresh Token (plain + hashed)
func GenerateRefreshToken() (string, string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}

	token := hex.EncodeToString(b)
	hash := sha256.Sum256([]byte(token))

	return token, hex.EncodeToString(hash[:]), nil
}

// 💾 Save Refresh Token
func SaveRefreshToken(db *gorm.DB, userID uint, hashedToken string, expiresAt time.Time) error {

	// delete old tokens (optional but good practice)
	if err := db.Where("user_id = ?", userID).Delete(&entity.RefreshToken{}).Error; err != nil {
		return err
	}

	refreshToken := entity.RefreshToken{
		UserId:    userID,
		Token:     hashedToken,
		ExpiredAt: expiresAt,
	}

	return db.Create(&refreshToken).Error
}

// ✅ Validate Refresh Token
func ValidateRefreshToken(db *gorm.DB, token string) (*entity.RefreshToken, error) {
	hash := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(hash[:])

	var retoken entity.RefreshToken

	err := db.Where("token = ? AND expired_at > ?", hashedToken, time.Now()).
		First(&retoken).Error

	if err != nil {
		return nil, errors.New("expired or invalid refresh token")
	}

	return &retoken, nil
}

// create a new access and refresh token

func RefreshAccessToken(db *gorm.DB, refreshToken string) (string, error) {

	// validate refresh token
	rt, err := ValidateRefreshToken(db, refreshToken)
	if err != nil {
		return "", err
	}

	// get user (you may need repo here)
	var user entity.User
	if err := db.First(&user, rt.UserId).Error; err != nil {
		return "", err
	}

	// generate new access token
	newAccessToken, err := GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return "", err
	}

	return newAccessToken, nil
}

// ❌ Delete Refresh Token
func DeleteReToken(db *gorm.DB, token string) error {
	hash := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(hash[:])

	return db.Where("token = ?", hashedToken).
		Delete(&entity.RefreshToken{}).Error
}