package seed

import (
	"backend/internal/config"
	"backend/internal/entity"
	"golang.org/x/crypto/bcrypt"
)

func SeedUsers() {
	db := config.DB

	users := []entity.User{
		{
			FullName:   "Admin User",
			Email:      "admin@example.com",
			Role:       "admin",
			IsVerified: true,
		},
		{
			FullName:   "Normal User",
			Email:      "user@example.com",
			Role:       "user",
			IsVerified: true,
		},
	}

	for _, u := range users {
		var existing entity.User
		if err := db.Where("email = ?", u.Email).First(&existing).Error; err == nil {
			continue // already exists
		}

		// hash password
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		u.HashPassword = string(hashedPassword)

		db.Create(&u)
	}
}
