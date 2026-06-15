package main

import (
	"backend/internal/config"
	"backend/internal/entity"
	"backend/internal/usecase"
	"fmt"
	"log"
)

func main() {
	// Initialize Config
	config.ConnectDB()
	db := config.DB

	if db == nil {
		log.Fatal("Could not connect to database")
	}

	var user entity.User
	if err := db.Where("email = ?", "support@gmail.com").First(&user).Error; err != nil {
		log.Fatalf("Failed to find support@gmail.com: %v", err)
	}

	fmt.Println("Found user:", user.Email, "Current role:", user.Role)

	if usecase.NormalizeRole(user.Role) != "supportagent" {
		user.Role = "supportagent"
		db.Model(&user).Update("role", "supportagent")
		fmt.Println("Updated user table role to supportagent")
	}

	var role entity.Role
	if err := db.Where("LOWER(name) = ?", "supportagent").First(&role).Error; err != nil {
		log.Fatalf("Failed to find supportagent role: %v", err)
	}

	db.Where("user_id = ?", user.ID).Delete(&entity.UserRole{})
	db.Create(&entity.UserRole{
		UserID:    user.ID,
		RoleID:    role.ID,
		IsPrimary: true,
	})

	fmt.Println("Successfully assigned supportagent role to support@gmail.com in user_roles table.")
}
