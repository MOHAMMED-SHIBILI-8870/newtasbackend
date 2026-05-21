package main

import (
	"context"
	"log"

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

	"google.golang.org/genai"
)

func main() {

	config.LoadEnv()
	config.ConnectDB()

	db := config.DB

	err := migrations.Migrations()
	if err != nil {
		panic(err)
	}
	seed.SeedUsers()
	apiKey := config.GetEnv("GEMINI_API_KEY", "")

	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY not found")
	}

	ctx := context.Background()

	geminiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})

	if err != nil {
		log.Fatalf("Failed to initialize Gemini Client: %v", err)
	}

	app := fiber.New()

	app.Use(logger.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173,http://localhost:5174",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowCredentials: true,
	}))

	// Repositories (Data Access Layer)
	
	tripRepo := repository.NewTripRepository(db)
	userRepo := repository.NewUserRepository(db)
	bookingRepo := repository.NewBookingRepository(db) // 👈 Added booking repo

	
	// Usecases (Business Logic Layer)
	
	tripUsecase := usecase.NewTripUsecase(tripRepo)
	adminUsecase := usecase.NewAdminUsecase(userRepo)
	// 👈 Initialize the booking business workflow with both required repos
	bookingUsecase := usecase.NewBookingUsecase(bookingRepo, tripRepo) 

	
	// Handlers (Controller Layer)
	
	tripHandler := handler.NewTripHandler(tripUsecase)
	adminHandler := handler.NewAdminHandler(adminUsecase)
	aiHandler := handler.NewAIHandler(geminiClient)
	bookingHandler := handler.NewBookingHandler(bookingUsecase) 

	
	// Routes Injection Mapping
	
	routes.AuthRoutes(app, db)
	routes.TripRoutes(app, tripHandler)
	routes.AdminRoutes(app, adminHandler)
	routes.SetupAIRoutes(app, aiHandler)
	routes.BookingRoutes(app, bookingHandler)

	// Start Server
	
	log.Println("Server running on port 8997")

	log.Fatal(app.Listen(":8997"))
}