package usecase

import (
	"backend/internal/entity"
	"context"
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
	"gorm.io/gorm/clause"
)

const (
	normalizedAdminRole   = "admin"
	normalizedAgencyRole  = "agency"
	normalizedGuideRole   = "guide"
	normalizedDriverRole  = "driver"
	normalizedSupportRole = "support"
	normalizedSupportAgentRole = "supportagent"
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
	case normalizedSupportAgentRole:
		return normalizedSupportAgentRole
	case normalizedUserRole:
		return normalizedUserRole
	default:
		return normalizedUserRole
	}
}

func IsValidRole(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case normalizedAdminRole, normalizedAgencyRole, normalizedGuideRole, normalizedDriverRole, normalizedSupportRole, normalizedSupportAgentRole, normalizedUserRole:
		return true
	default:
		return false
	}
}

func getJWTSecret() (string, error) {
	if secretKey := os.Getenv("JWT_SECRET"); secretKey != "" {
		return secretKey, nil
	}

	if secretKey := os.Getenv("JWT_SECRETKEY"); secretKey != "" {
		return secretKey, nil
	}

	return "", errors.New("JWT secret not configured")
}

// Generate Access Token
func GenerateAccessToken(userID uint, role string) (string, error) {
	secretKey, err := getJWTSecret()
	if err != nil {
		return "", err
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
func SaveRefreshToken(db *gorm.DB, ctx context.Context, userID uint, hashedToken string, expiresAt time.Time) error {
	if db == nil {
		return errors.New("database unavailable")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Keep refresh tokens single-session per user so a new login invalidates older sessions.
		if err := tx.Where("user_id = ?", userID).Delete(&entity.RefreshToken{}).Error; err != nil {
			return err
		}

		refreshToken := entity.RefreshToken{
			UserID:    userID,
			Token:     hashedToken,
			ExpiredAt: expiresAt,
		}

		return tx.Create(&refreshToken).Error
	})
}

// ✅ Validate Refresh Token
func ValidateRefreshToken(db *gorm.DB, ctx context.Context, token string) (*entity.RefreshToken, error) {
	if db == nil {
		return nil, errors.New("database unavailable")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	hash := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(hash[:])

	var retoken entity.RefreshToken

	err := db.WithContext(ctx).Where("token = ? AND expired_at > ?", hashedToken, time.Now()).
		First(&retoken).Error

	if err != nil {
		return nil, errors.New("expired or invalid refresh token")
	}

	return &retoken, nil
}

// RevokeRefreshTokensByUserID removes every refresh token tied to one account.
func RevokeRefreshTokensByUserID(db *gorm.DB, ctx context.Context, userID uint) error {
	if db == nil {
		return errors.New("database unavailable")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	return db.WithContext(ctx).Where("user_id = ?", userID).Delete(&entity.RefreshToken{}).Error
}

// RotateRefreshToken consumes the old refresh token and issues a replacement atomically.
func RotateRefreshToken(db *gorm.DB, ctx context.Context, refreshToken string) (*entity.User, string, time.Time, error) {
	if db == nil {
		return nil, "", time.Time{}, errors.New("database unavailable")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var user entity.User
	var newPlainToken string
	var expiresAt time.Time

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		hash := sha256.Sum256([]byte(refreshToken))
		hashedToken := hex.EncodeToString(hash[:])

		var current entity.RefreshToken
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("token = ? AND expired_at > ?", hashedToken, time.Now()).
			First(&current).Error; err != nil {
			return errors.New("expired or invalid refresh token")
		}

		if err := tx.First(&user, current.UserID).Error; err != nil {
			return err
		}
		if user.IsBlocked {
			return errors.New("account is blocked")
		}
		if !user.IsVerified {
			return errors.New("account is not verified")
		}

		if err := tx.Where("user_id = ?", user.ID).Delete(&entity.RefreshToken{}).Error; err != nil {
			return err
		}

		plainToken, hashedNewToken, err := GenerateRefreshToken()
		if err != nil {
			return err
		}

		expiresAt = time.Now().Add(7 * 24 * time.Hour)
		if err := tx.Create(&entity.RefreshToken{
			UserID:    user.ID,
			Token:     hashedNewToken,
			ExpiredAt: expiresAt,
		}).Error; err != nil {
			return err
		}

		newPlainToken = plainToken
		return nil
	})
	if err != nil {
		return nil, "", time.Time{}, err
	}

	return &user, newPlainToken, expiresAt, nil
}

// create a new access and refresh token

func RefreshAccessToken(db *gorm.DB, ctx context.Context, refreshToken string) (string, error) {
	if db == nil {
		return "", errors.New("database unavailable")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	// validate refresh token
	rt, err := ValidateRefreshToken(db, ctx, refreshToken)
	if err != nil {
		return "", err
	}

	// get user (you may need repo here)
	var user entity.User
	if err := db.WithContext(ctx).First(&user, rt.UserID).Error; err != nil {
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
func DeleteReToken(db *gorm.DB, ctx context.Context, token string) error {
	if db == nil {
		return errors.New("database unavailable")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	hash := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(hash[:])

	return db.WithContext(ctx).Where("token = ?", hashedToken).
		Delete(&entity.RefreshToken{}).Error
}
