package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"

	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/handler/routes"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/response"
	"backend/internal/seed"
	"backend/internal/service"
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

	if err := validateStartupEnv(); err != nil {
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
	seed.SeedRBAC()

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
	tripSlotRepo := repository.NewTripSlotRepository(db)
	userRepo := repository.NewUserRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	aiTripRequestRepo := repository.NewAITripRequestRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)
	vehicleRepo := repository.NewVehicleRepository(db)
	offerRepo := repository.NewOfferRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	complaintRepo := repository.NewComplaintRepository(db)
	trackingRepo := repository.NewTrackingRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	razorpayService := service.NewRazorpayService()

	notificationUsecase := usecase.NewNotificationUsecase(notificationRepo)
	tripUsecase := usecase.NewTripUsecase(tripRepo)
	tripPlanUsecase := usecase.NewTripPlanUsecase(tripPlanRepo)
	tripSlotUsecase := usecase.NewTripSlotUsecase(tripSlotRepo, tripRepo, vehicleRepo, db)
	adminUsecase := usecase.NewAdminUsecase(userRepo, roleRepo, db)
	roleUsecase := usecase.NewRoleUsecase(roleRepo, userRepo, db)
	permissionUsecase := usecase.NewPermissionUsecase(permissionRepo, roleRepo, userRepo)
	vehicleUsecase := usecase.NewVehicleUsecase(vehicleRepo, tripRepo, userRepo, db)
	offerUsecase := usecase.NewOfferUsecase(offerRepo)
	reviewUsecase := usecase.NewReviewUsecase(reviewRepo, bookingRepo, tripRepo, db)
	complaintUsecase := usecase.NewComplaintUsecase(complaintRepo, bookingRepo)
	trackingUsecase := usecase.NewTrackingUsecase(trackingRepo, bookingRepo, vehicleRepo)
	paymentUsecase := usecase.NewPaymentUsecase(paymentRepo,bookingRepo,razorpayService)
	bookingUsecase := usecase.NewBookingUsecase(bookingRepo, tripRepo, userRepo, offerRepo, db, notificationUsecase, tripSlotRepo)
	aiTripRequestUsecase := usecase.NewAITripRequestUsecase(aiTripRequestRepo, tripRepo, userRepo, notificationUsecase, db)

	tripHandler := handler.NewTripHandler(tripUsecase)
	tripPlanHandler := handler.NewTripPlanHandler(tripPlanUsecase)
	tripSlotHandler := handler.NewTripSlotHandler(tripSlotUsecase)
	adminHandler := handler.NewAdminHandler(adminUsecase)
	ragUsecase := usecase.NewRAGUsecase(tripRepo)
	aiHandler := handler.NewAIHandler(geminiClient,aiTripRequestUsecase,ragUsecase)
	bookingHandler := handler.NewBookingHandler(bookingUsecase)
	notificationHandler := handler.NewNotificationHandler(notificationUsecase)
	roleHandler := handler.NewRoleHandler(roleUsecase, permissionUsecase)
	permissionHandler := handler.NewPermissionHandler(permissionUsecase)
	vehicleHandler := handler.NewVehicleHandler(vehicleUsecase)
	offerHandler := handler.NewOfferHandler(offerUsecase)
	reviewHandler := handler.NewReviewHandler(reviewUsecase)
	complaintHandler := handler.NewComplaintHandler(complaintUsecase)
	trackingHandler := handler.NewTrackingHandler(trackingUsecase)
	paymentHandler := handler.NewPaymentHandler(paymentUsecase)

	authMiddleware := middleware.AuthMiddleware(userRepo)
	adminMiddleware := middleware.RoleMiddleware("admin")
	permissionMiddleware := middleware.PermissionMiddleware(permissionUsecase)

	routes.AuthRoutes(app, db)
	routes.TripRoutes(app, tripHandler, authMiddleware, adminMiddleware)
	routes.TripSlotRoutes(app, tripSlotHandler, authMiddleware, adminMiddleware)
	routes.TripPlanRoutes(app, tripPlanHandler, authMiddleware, adminMiddleware)
	routes.AdminRoutes(app, adminHandler, authMiddleware, adminMiddleware)
	routes.SetupAIRoutes(app, aiHandler, authMiddleware, adminMiddleware)
	routes.BookingRoutes(app, bookingHandler, authMiddleware, adminMiddleware)
	routes.NotificationRoutes(app, notificationHandler, authMiddleware, adminMiddleware)
	routes.RBACRoutes(app, roleHandler, permissionHandler, authMiddleware, permissionMiddleware)
	routes.VehicleRoutes(app, vehicleHandler, authMiddleware, permissionMiddleware)
	routes.OfferRoutes(app, offerHandler, authMiddleware, permissionMiddleware)
	routes.ReviewRoutes(app, reviewHandler, authMiddleware, permissionMiddleware)
	routes.ComplaintRoutes(app, complaintHandler, authMiddleware, permissionMiddleware)
	routes.TrackingRoutes(app, trackingHandler, authMiddleware, permissionMiddleware)
	routes.PaymentRoutes(app,paymentHandler,authMiddleware)

	port := config.GetEnv("PORT", "8997")
	log.Printf("Server running on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

func validateStartupEnv() error {
	required := []string{
		"DB_HOST",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"DB_PORT",
		"SMTP_HOST",
		"SMTP_PORT",
		"EMAIL_FROM",
		"EMAIL_PASSWORD",
		"PORT",

		"RAZORPAY_KEY_ID",
		"RAZORPAY_KEY_SECRET",
	}

	if err := config.ValidateRequiredEnv(required...); err != nil {
		return err
	}

	if strings.TrimSpace(os.Getenv("JWT_SECRET")) == "" && strings.TrimSpace(os.Getenv("JWT_SECRETKEY")) == "" {
		return errors.New("missing required environment variable: JWT_SECRET")
	}

	return nil
}
