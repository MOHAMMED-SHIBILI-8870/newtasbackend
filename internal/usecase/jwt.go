package usecase

import (
	"backend/internal/entity"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

const (
	normalizedAdminRole   = "admin"
	normalizedAgencyRole  = "agency"
	normalizedGuideRole   = "guide"
	normalizedDriverRole  = "driver"
	normalizedSupportRole = "support"
	normalizedUserRole    = "user"
)

type AuthClaims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NormalizeRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case normalizedAdminRole:
		return normalizedAdminRole
	case normalizedAgencyRole:
		return normalizedAgencyRole
	case normalizedGuideRole:
		return normalizedGuideRole
	case normalizedDriverRole:
		return normalizedDriverRole
	case normalizedSupportRole:
		return normalizedSupportRole
	case normalizedUserRole:
		return normalizedUserRole
	default:
		return normalizedUserRole
	}
}

func IsValidRole(role string) bool {
	switch NormalizeRole(role) {
	case normalizedAdminRole, normalizedAgencyRole, normalizedGuideRole, normalizedDriverRole, normalizedSupportRole, normalizedUserRole:
		return true
	default:
		return false
	}
}

// Generate Access Token
func GenerateAccessToken(userID uint, role string) (string, error) {
	secretKey := os.Getenv("JWT_SECRETKEY")
	if secretKey == "" {
		return "", errors.New("JWT secret key not set")
	}

	now := time.Now().UTC()
	claims := AuthClaims{
		UserID: userID,
		Role:   NormalizeRole(role),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(userID), 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
		},
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
		UserID:    userID,
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
	if err := db.First(&user, rt.UserID).Error; err != nil {
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
