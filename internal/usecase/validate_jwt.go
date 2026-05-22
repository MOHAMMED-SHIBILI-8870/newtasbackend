package usecase

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func ValidateJwt(tokenStr string) (uint, string, error) {

	secretKey := os.Getenv("JWT_SECRETKEY")
	if secretKey == "" {
		return 0, "", errors.New("JWT secret not configured")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &AuthClaims{}, func(token *jwt.Token) (any, error) {

		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, jwt.ErrTokenSignatureInvalid
		}

		return []byte(secretKey), nil
	})

	if err != nil {
		return 0, "", err
	}

	if token == nil || !token.Valid {
		return 0, "", jwt.ErrTokenInvalidClaims
	}

	claims, ok := token.Claims.(*AuthClaims)
	if !ok {
		return 0, "", jwt.ErrTokenInvalidClaims
	}

	var userID uint
	if claims.UserID != 0 {
		userID = claims.UserID
	} else if claims.Subject != "" {
		id, err := strconv.Atoi(claims.Subject)
		if err != nil {
			return 0, "", err
		}
		userID = uint(id)
	}

	if userID == 0 {
		return 0, "", jwt.ErrTokenInvalidClaims
	}

	role := NormalizeRole(strings.TrimSpace(claims.Role))

	return userID, role, nil
}
