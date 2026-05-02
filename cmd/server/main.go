package main

import (
	"backend/internal/config"
	"backend/internal/handler/routes"
	"backend/migrations"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	config.ConnectDB()

	migrations.Migrations()

	app := fiber.New()
	app.Use(cors.New(cors.Config{
    AllowOrigins: "http://localhost:5173,http://localhost:5174",
    AllowHeaders: "Origin, Content-Type, Accept, Authorization",
    AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
	AllowCredentials: true,
}))
	routes.AuthRoutes(app)
	app.Listen(":8997")
}
