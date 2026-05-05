package main

import (
	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/handler/routes"
	"backend/internal/repository"
	"backend/internal/seed"
	"backend/internal/usecase"
	"backend/migrations"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// 1. Database Connection & Migrations
	config.ConnectDB()
	db := config.DB // Assuming your config package exports the GORM DB instance
	migrations.Migrations()
	seed.SeedUsers()

	app := fiber.New()

	// 2. Middleware
	app.Use(logger.New()) // This will print every request to your terminal
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173,http://localhost:5174",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowCredentials: true,
	}))

	// trip
	tripRepo := repository.NewTripRepository(db)
	tripUsecase := usecase.NewTripUsecase(tripRepo)
	tripHandler := handler.NewTripHandler(tripUsecase)
	// Admin
	userRepo := repository.NewUserRepository(db)         
    adminUsecase := usecase.NewAdminUsecase(userRepo)
    adminHandler := handler.NewAdminHandler(adminUsecase)

	// 4. Register Routes
	routes.AuthRoutes(app,db)
	routes.TripRoutes(app, tripHandler) 
	routes.AdminRoutes(app, adminHandler)

	// 5. Start Server
	app.Listen(":8997")
}
