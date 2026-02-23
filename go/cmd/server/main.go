// @title DSView API Go Service
// @version 1.0.0-alpha
// @description This is the backend API for DSView application In Services GO
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url https://github.com/Project-DSView/backend
// @contact.email 65070209@kmitl.ac.th

// @host localhost/go
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name dsview-api-key
// @description API key authentication for secure endpoints.
package main

import (
	"fmt"
	"os"

	_ "github.com/Project-DSView/backend/go/docs"
	"github.com/Project-DSView/backend/go/internal/api/routes"
	"github.com/Project-DSView/backend/go/internal/infrastructure/config"
	"github.com/Project-DSView/backend/go/internal/infrastructure/setup"
	"github.com/Project-DSView/backend/go/pkg/logger"
)

func main() {
	// Load configuration
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	cfg, err := config.Load(env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize structured logger early (before DB setup)
	appLogger, logErr := logger.NewLogger("go-app", logger.INFO, "/app/logs")
	if logErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to initialize file logger: %v\n", logErr)
		// Fallback: create logger writing to current dir
		appLogger, _ = logger.NewLogger("go-app", logger.INFO, "./logs")
	}
	if appLogger != nil {
		logger.SetGlobalLogger(appLogger)
	}

	// Setup database
	db, err := setup.SetupDatabase(cfg)
	if err != nil {
		logger.Fatal("Failed to setup database", err)
	}

	// Setup all services
	logger.Info("Setting up core services...")
	services, err := setup.SetupCoreServices(db, cfg)
	if err != nil {
		logger.Fatal("Failed to setup services", err)
	}
	logger.Info("Core services setup completed")

	// Setup routes
	app := routes.SetupRoutes(
		cfg,
		services.OAuthService,
		services.JWTService,
		services.UserService,
		services.CourseService,
		services.TestCaseService,
		services.EnrollmentService,
		services.InvitationService,
		services.SubmissionService,
		services.ProgressService,
		services.DraftService,
		services.QueueService,
		services.StorageService,
		services.DeadlineCheckerService,
		services.CourseScoreService,
		services.CourseMaterialService,
		services.DB,
	)

	// Start server
	serverAddr := cfg.Server.Host + ":" + cfg.Server.Port
	logger.Infof("Server starting on %s", serverAddr)
	logger.Infof("Connected to database: %s", cfg.Database.DBName)
	logger.Infof("MinIO service initialized for bucket: %s", cfg.MinIO.BucketName)

	// // Check if HTTPS certificates exist and start appropriate server
	// if _, err := os.Stat("./localhost.pem"); err == nil {
	// 	if _, err := os.Stat("./localhost-key.pem"); err == nil {
	// 		logger.Infof("HTTPS certificates found, starting HTTPS server on %s", serverAddr)
	// 		logger.Fatal("Server failed", app.ListenTLS(serverAddr, "./localhost.pem", "./localhost-key.pem"))
	// 	}
	// }

	// Fallback to HTTP if no HTTPS certificates
	logger.Infof("No HTTPS certificates found, starting HTTP server on %s", serverAddr)
	logger.Fatal("Server failed", app.Listen(serverAddr))
}
