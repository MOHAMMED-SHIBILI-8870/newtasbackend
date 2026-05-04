package usecase

import (
	"fmt"
	"os"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

func ValidateJwt(tokenStr string) (uint, string, error) {

	secretKey := os.Getenv("JWT_SECRETKEY")
	fmt.Println("SECRET:", secretKey)
	if secretKey == "" {
		return 0, "", fmt.Errorf("JWT secret not configured")
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {

		// safer validation
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method")
		}

		return []byte(secretKey), nil
	})

	if err != nil {
		return 0, "", err
	}

	if token == nil || !token.Valid {
		return 0, "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", fmt.Errorf("invalid claims")
	}

	// user_id safe conversion
	var userID uint
	switch v := claims["user_id"].(type) {
	case float64:
		userID = uint(v)
	case string:
		id, err := strconv.Atoi(v)
		if err != nil {
			return 0, "", fmt.Errorf("invalid userId in token")
		}
		userID = uint(id)
	default:
		return 0, "", fmt.Errorf("invalid userId in token")
	}

	// role check
	roleVal, exists := claims["role"]
	if !exists {
		return 0, "", fmt.Errorf("role missing in token")
	}

	role, ok := roleVal.(string)
	if !ok {
		return 0, "", fmt.Errorf("invalid role in token")
	}

	return userID, role, nil
}