package main

import (
	"backend/internal/config"
	"backend/internal/entity"

	"github.com/gofiber/fiber/v2"
)

func main() {
	config.ConnectDB()

	// ✅ migrate tables
	config.DB.AutoMigrate(&entity.User{})

	app := fiber.New()
	app.Listen(":8997")
}