package main

import (
	"context"
	"log"

	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/handler/routes"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/response"
	"backend/internal/seed"
	"backend/internal/usecase"
	"backend/migrations"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"google.golang.org/genai"
)

func main() {
	if err := config.LoadEnv(); err != nil {
		log.Printf("env load warning: %v", err)
	}

	if err := config.ValidateRequiredEnv("DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_PORT", "JWT_SECRETKEY"); err != nil {
		log.Fatal(err)
	}

	if err := config.ConnectDB(); err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	db := config.DB

	if err := migrations.Migrations(); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	seed.SeedUsers()

	var geminiClient *genai.Client
	apiKey := config.GetEnv("GEMINI_API_KEY", "")
	if apiKey != "" {
		ctx := context.Background()
		client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
		if err != nil {
			log.Printf("failed to initialize Gemini client, using fallback AI generator: %v", err)
		} else {
			geminiClient = client
		}
	} else {
		log.Println("GEMINI_API_KEY not set, using fallback AI generator")
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: response.FiberErrorHandler,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     config.GetEnv("CORS_ORIGINS", "http://localhost:5173,http://localhost:5174,http://127.0.0.1:5173,http://127.0.0.1:5174"),
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowCredentials: true,
	}))

	tripRepo := repository.NewTripRepository(db)
	tripPlanRepo := repository.NewTripPlanRepository(db)
	userRepo := repository.NewUserRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	aiTripRequestRepo := repository.NewAITripRequestRepository(db)

	notificationUsecase := usecase.NewNotificationUsecase(notificationRepo)
	tripUsecase := usecase.NewTripUsecase(tripRepo)
	tripPlanUsecase := usecase.NewTripPlanUsecase(tripPlanRepo)
	adminUsecase := usecase.NewAdminUsecase(userRepo)
	bookingUsecase := usecase.NewBookingUsecase(bookingRepo, tripRepo, notificationUsecase)
	aiTripRequestUsecase := usecase.NewAITripRequestUsecase(aiTripRequestRepo, tripRepo, userRepo, notificationUsecase)

	tripHandler := handler.NewTripHandler(tripUsecase)
	tripPlanHandler := handler.NewTripPlanHandler(tripPlanUsecase)
	adminHandler := handler.NewAdminHandler(adminUsecase)
	aiHandler := handler.NewAIHandler(geminiClient, aiTripRequestUsecase)
	bookingHandler := handler.NewBookingHandler(bookingUsecase)
	notificationHandler := handler.NewNotificationHandler(notificationUsecase)

	authMiddleware := middleware.AuthMiddleware(userRepo)
	adminMiddleware := middleware.RoleMiddleware("admin")

	routes.AuthRoutes(app, db)
	routes.TripRoutes(app, tripHandler, authMiddleware, adminMiddleware)
	routes.TripPlanRoutes(app, tripPlanHandler, authMiddleware, adminMiddleware)
	routes.AdminRoutes(app, adminHandler, authMiddleware, adminMiddleware)
	routes.SetupAIRoutes(app, aiHandler, authMiddleware, adminMiddleware)
	routes.BookingRoutes(app, bookingHandler, authMiddleware, adminMiddleware)
	routes.NotificationRoutes(app, notificationHandler, authMiddleware, adminMiddleware)

	log.Println("Server running on port 8997")
	log.Fatal(app.Listen(":8997"))
}
