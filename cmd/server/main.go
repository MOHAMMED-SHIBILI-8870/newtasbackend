package main

import (
	"backend/internal/config"
	"backend/internal/handler/routes"
	"backend/migrations"

	"github.com/gofiber/fiber/v2"
)

func main() {
	config.ConnectDB()

	migrations.Migrations()

	app := fiber.New()
	routes.AuthRoutes(app)
	app.Listen(":8997")
}
